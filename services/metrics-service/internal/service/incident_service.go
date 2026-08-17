package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/modulops/metrics-service/internal/configstore"
)

// AlertmanagerPayload est le corps envoye par Alertmanager a un receiver de
// type webhook.
type AlertmanagerPayload struct {
	Status      string              `json:"status"`
	Receiver    string              `json:"receiver"`
	GroupLabels map[string]string   `json:"groupLabels"`
	Alerts      []AlertmanagerAlert `json:"alerts"`
}

// AlertmanagerAlert decrit une alerte du lot.
type AlertmanagerAlert struct {
	Status      string            `json:"status"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    time.Time         `json:"startsAt"`
}

// IncidentService ouvre une issue Forgejo a la reception d'une alerte.
type IncidentService struct {
	cfgStore   *configstore.Client
	httpClient *http.Client
}

// NewIncidentService cree le service.
func NewIncidentService(authServiceURL string) *IncidentService {
	return &IncidentService{
		cfgStore:   configstore.New(authServiceURL, "metrics"),
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// HandleAlert ouvre une issue pour un lot d'alertes actives.
//
// Les alertes resolues sont ignorees : Alertmanager rappelle le lot a la
// resolution, ouvrir une issue a ce moment creerait un doublon sans objet.
func (s *IncidentService) HandleAlert(ctx context.Context, payload *AlertmanagerPayload) (string, error) {
	if payload.Status == "resolved" {
		return "", nil
	}
	actives := make([]AlertmanagerAlert, 0, len(payload.Alerts))
	for _, alerte := range payload.Alerts {
		if alerte.Status != "resolved" {
			actives = append(actives, alerte)
		}
	}
	if len(actives) == 0 {
		return "", nil
	}

	depot, err := s.cfgStore.GetShared(ctx, "incident_repository")
	if err != nil || depot == "" {
		return "", fmt.Errorf("depot des incidents non configure (cle incident_repository)")
	}
	// Jeton dedie : la creation d'issues demande write:issue, que le jeton de
	// lecture des modules n'a pas. Repli sur forgejo_token si non configure.
	token, _ := s.cfgStore.GetShared(ctx, "incident_token")
	if token == "" {
		token, _ = s.cfgStore.GetShared(ctx, "forgejo_token")
	}
	if token == "" {
		return "", fmt.Errorf("jeton Forgejo absent")
	}
	baseURL, _ := s.cfgStore.GetShared(ctx, "forgejo_url")
	if baseURL == "" {
		baseURL = "https://codeberg.org"
	}

	titre, corps := redigerIssue(payload, actives)

	body, err := json.Marshal(map[string]interface{}{
		"title":  titre,
		"body":   corps,
		"labels": []int{},
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/api/v1/repos/%s/issues", strings.TrimSuffix(baseURL, "/"), depot)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("creation de l'issue refusee (HTTP %d)", resp.StatusCode)
	}

	var cree struct {
		Number int    `json:"number"`
		URL    string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cree); err != nil {
		return "", err
	}

	s.poserEtiquettes(ctx, baseURL, depot, token, cree.Number)
	return cree.URL, nil
}

// poserEtiquettes applique incident et a-qualifier a l'issue creee.
//
// Les etiquettes sont posees dans un second appel : l'API attend des
// identifiants numeriques, qu'il faut resoudre depuis leurs noms.
func (s *IncidentService) poserEtiquettes(ctx context.Context, baseURL, depot, token string, numero int) {
	voulues := map[string]bool{"incident": true, "a-qualifier": true}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/v1/repos/%s/labels", strings.TrimSuffix(baseURL, "/"), depot), nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "token "+token)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var etiquettes []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&etiquettes); err != nil {
		return
	}

	ids := make([]int, 0, 2)
	for _, e := range etiquettes {
		if voulues[e.Name] {
			ids = append(ids, e.ID)
		}
	}
	if len(ids) == 0 {
		return
	}

	body, _ := json.Marshal(map[string]interface{}{"labels": ids})
	poser, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v1/repos/%s/issues/%d/labels", strings.TrimSuffix(baseURL, "/"), depot, numero),
		bytes.NewReader(body))
	if err != nil {
		return
	}
	poser.Header.Set("Authorization", "token "+token)
	poser.Header.Set("Content-Type", "application/json")
	if r, err := s.httpClient.Do(poser); err == nil {
		r.Body.Close()
	}
}

// redigerIssue produit le titre et le corps a partir du gabarit d'incident.
func redigerIssue(payload *AlertmanagerPayload, actives []AlertmanagerAlert) (string, string) {
	premiere := actives[0]
	nom := premiere.Labels["alertname"]
	if nom == "" {
		nom = "Alerte"
	}
	severite := premiere.Labels["severity"]

	titre := fmt.Sprintf("[%s] %s", strings.ToUpper(severite), nom)
	if len(actives) > 1 {
		titre = fmt.Sprintf("%s (%d alertes)", titre, len(actives))
	}

	services := make([]string, 0, len(actives))
	vus := map[string]bool{}
	for _, a := range actives {
		s := a.Labels["job"]
		if s == "" {
			s = a.Labels["service"]
		}
		if s != "" && !vus[s] {
			vus[s] = true
			services = append(services, s)
		}
	}
	sort.Strings(services)

	detection := premiere.StartsAt
	if detection.IsZero() {
		detection = time.Now()
	}

	var b strings.Builder
	b.WriteString("## Detection\n\n")
	fmt.Fprintf(&b, "%s, alerte `%s` levee par vmalert.\n\n", detection.Format("02/01/2006 15h04"), nom)

	b.WriteString("## Perimetre touche\n\n")
	if len(services) > 0 {
		fmt.Fprintf(&b, "%s\n\n", strings.Join(services, ", "))
	} else {
		b.WriteString("A preciser.\n\n")
	}

	b.WriteString("## Severite proposee\n\n")
	fmt.Fprintf(&b, "%s (a confirmer a la qualification)\n\n", severite)

	b.WriteString("## Chronologie\n\n")
	for _, a := range actives {
		description := a.Annotations["description"]
		if description == "" {
			description = a.Annotations["summary"]
		}
		fmt.Fprintf(&b, "- %s : %s\n", a.StartsAt.Format("15h04"), description)
	}
	b.WriteString("\n## Impact utilisateurs\n\nA completer par l'intervenant.\n\n")
	b.WriteString("## Actions engagees\n\nA completer par l'intervenant.\n")

	return titre, b.String()
}
