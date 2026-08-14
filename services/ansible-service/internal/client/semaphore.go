// Package client fournit le client REST vers l'API Ansible Semaphore.
//
// Semaphore expose tout sous /api/project/{project_id}/... ; les réponses sont
// traduites au format AWX historique pour que le frontend ne nécessite aucun
// changement.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// statusMap traduit les statuts Semaphore vers les statuts AWX attendus par le frontend.
var statusMap = map[string]string{
	"waiting": "pending",
	"running": "running",
	"success": "successful",
	"error":   "failed",
	"stopped": "canceled",
}

// Paginated est la forme de réponse paginée AWX consommée par le frontend.
type Paginated struct {
	Count    int              `json:"count"`
	Results  []map[string]any `json:"results"`
	Next     *string          `json:"next"`
	Previous *string          `json:"previous"`
}

// EmptyPaginated retourne une réponse paginée vide.
func EmptyPaginated() *Paginated {
	return &Paginated{Count: 0, Results: []map[string]any{}}
}

// Semaphore est le client de l'API REST Ansible Semaphore.
type Semaphore struct {
	baseURL    string
	apiToken   string
	projectID  int
	httpClient *http.Client

	// Nom du projet mémoïsé : constant pour un projectID donné, et le client
	// est reconstruit à chaque changement de configuration.
	nameOnce   sync.Once
	nameCached string
}

// NewSemaphore construit un client Semaphore.
func NewSemaphore(baseURL, apiToken string, projectID int) *Semaphore {
	return &Semaphore{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiToken:   apiToken,
		projectID:  projectID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ProjectID retourne l'identifiant du projet Semaphore courant.
func (s *Semaphore) ProjectID() int { return s.projectID }

// Configured indique si le client dispose d'une URL Semaphore.
func (s *Semaphore) Configured() bool { return s.baseURL != "" }

// request exécute une requête sur /api{path} et décode le JSON dans out.
// Retourne (false, nil) si le client n'est pas configuré ou si Semaphore
// répond une erreur HTTP (comportement héritier du service Python : les
// erreurs amont sont loguées et traduites en absence de données).
func (s *Semaphore) request(ctx context.Context, method, path string, payload any, out any) (bool, error) {
	if s.baseURL == "" {
		return false, nil
	}
	url := s.baseURL + "/api" + path

	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return false, err
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if s.apiToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.apiToken)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Printf("Semaphore: erreur requête %s %s: %v", method, url, err)
		return false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		log.Printf("Semaphore: HTTP %d %s: %s", resp.StatusCode, url, string(b))
		return false, nil
	}
	if resp.StatusCode == http.StatusNoContent || out == nil {
		return true, nil
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if len(bytes.TrimSpace(raw)) == 0 {
		return true, nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return false, fmt.Errorf("Semaphore: réponse JSON invalide %s: %w", url, err)
	}
	return true, nil
}

// ── Projets ──────────────────────────────────────────────────────────────────

// GetProjects liste les projets Semaphore au format AWX.
func (s *Semaphore) GetProjects(ctx context.Context) (*Paginated, error) {
	var raw []map[string]any
	ok, err := s.request(ctx, http.MethodGet, "/projects", nil, &raw)
	if err != nil || !ok {
		return nil, err
	}
	results := make([]map[string]any, 0, len(raw))
	for _, p := range raw {
		results = append(results, s.mapProject(p))
	}
	return &Paginated{Count: len(results), Results: results}, nil
}

// GetProject retourne un projet par id.
func (s *Semaphore) GetProject(ctx context.Context, projectID int) (map[string]any, error) {
	var raw map[string]any
	ok, err := s.request(ctx, http.MethodGet, fmt.Sprintf("/project/%d", projectID), nil, &raw)
	if err != nil || !ok {
		return nil, err
	}
	return s.mapProject(raw), nil
}

func (s *Semaphore) mapProject(p map[string]any) map[string]any {
	return map[string]any{
		"id":          p["id"],
		"name":        str(p["name"]),
		"description": str(p["description"]),
		"scm_type":    "git",
	}
}

// ── Jobs (tasks Semaphore) ───────────────────────────────────────────────────

// GetJobs liste les tâches du projet, traduites au format job AWX.
func (s *Semaphore) GetJobs(ctx context.Context, pageSize int) (*Paginated, error) {
	var raw []map[string]any
	path := fmt.Sprintf("/project/%d/tasks?sort=id&order=desc&limit=%d", s.projectID, pageSize)
	ok, err := s.request(ctx, http.MethodGet, path, nil, &raw)
	if err != nil || !ok {
		return nil, err
	}
	templates := s.templatesByID(ctx)
	inventories := s.inventoriesByID(ctx)
	projectName := s.projectName(ctx)
	results := make([]map[string]any, 0, len(raw))
	for _, t := range raw {
		results = append(results, s.mapTask(t, templates, inventories, projectName))
	}
	return &Paginated{Count: len(results), Results: results}, nil
}

// GetJob retourne une tâche par id, au format job AWX. Contrairement aux
// listes, l'enrichissement ne charge que le template et l'inventaire
// référencés par la tâche, pas les listes complètes.
func (s *Semaphore) GetJob(ctx context.Context, jobID int) (map[string]any, error) {
	var raw map[string]any
	path := fmt.Sprintf("/project/%d/tasks/%d", s.projectID, jobID)
	ok, err := s.request(ctx, http.MethodGet, path, nil, &raw)
	if err != nil || !ok {
		return nil, err
	}

	templates := map[int]map[string]any{}
	inventories := map[int]map[string]any{}
	if templateID := intVal(raw["template_id"]); templateID != 0 {
		if template := s.rawTemplate(ctx, templateID); template != nil {
			templates[templateID] = template
			if inventoryID := intVal(template["inventory_id"]); inventoryID != 0 {
				if inventory := s.rawInventory(ctx, inventoryID); inventory != nil {
					inventories[inventoryID] = inventory
				}
			}
		}
	}
	return s.mapTask(raw, templates, inventories, s.projectName(ctx)), nil
}

// rawTemplate retourne un template Semaphore brut (non traduit) par id.
func (s *Semaphore) rawTemplate(ctx context.Context, templateID int) map[string]any {
	var raw map[string]any
	path := fmt.Sprintf("/project/%d/templates/%d", s.projectID, templateID)
	if ok, err := s.request(ctx, http.MethodGet, path, nil, &raw); err != nil || !ok {
		return nil
	}
	return raw
}

// rawInventory retourne un inventaire Semaphore brut (non traduit) par id.
func (s *Semaphore) rawInventory(ctx context.Context, inventoryID int) map[string]any {
	var raw map[string]any
	path := fmt.Sprintf("/project/%d/inventory/%d", s.projectID, inventoryID)
	if ok, err := s.request(ctx, http.MethodGet, path, nil, &raw); err != nil || !ok {
		return nil
	}
	return raw
}

// GetJobStdout retourne la sortie agrégée d'une tâche.
func (s *Semaphore) GetJobStdout(ctx context.Context, jobID int) (string, error) {
	var raw []map[string]any
	path := fmt.Sprintf("/project/%d/tasks/%d/output", s.projectID, jobID)
	ok, err := s.request(ctx, http.MethodGet, path, nil, &raw)
	if err != nil || !ok {
		return "", err
	}
	lines := make([]string, 0, len(raw))
	for _, entry := range raw {
		lines = append(lines, str(entry["output"]))
	}
	return strings.Join(lines, "\n"), nil
}

func (s *Semaphore) mapTask(t map[string]any, templates, inventories map[int]map[string]any, projectName string) map[string]any {
	status := str(t["status"])
	if mapped, found := statusMap[status]; found {
		status = mapped
	} else if status == "" {
		status = "unknown"
	}

	template := templates[intVal(t["template_id"])]
	templateName := str(template["alias"])
	if templateName == "" {
		templateName = str(template["name"])
	}

	var inventoryName any
	inventoryID, hasInventory := template["inventory_id"]
	if hasInventory {
		if inv, found := inventories[intVal(inventoryID)]; found {
			inventoryName = inv["name"]
		}
	}

	name := templateName
	if name == "" {
		name = fmt.Sprintf("Task #%v", t["id"])
	}

	elapsed := floatVal(t["duration"])
	if elapsed == 0 {
		elapsed = computeElapsed(str(t["created"]), str(t["end"]))
	}

	summary := map[string]any{"job_template": nil, "inventory": nil, "project": nil}
	if templateName != "" {
		summary["job_template"] = map[string]any{"name": templateName}
	}
	if inventoryName != nil {
		summary["inventory"] = map[string]any{"name": inventoryName}
	}
	if projectName != "" {
		summary["project"] = map[string]any{"name": projectName}
	}

	return map[string]any{
		"id":                t["id"],
		"name":              name,
		"status":            status,
		"started":           t["created"],
		"finished":          t["end"],
		"elapsed":           elapsed,
		"job_template":      t["template_id"],
		"job_template_name": templateName,
		"inventory":         inventoryID,
		"inventory_name":    inventoryName,
		"project":           s.projectID,
		"project_name":      projectName,
		"summary_fields":    summary,
	}
}

// ── Templates ────────────────────────────────────────────────────────────────

// GetJobTemplates liste les templates du projet au format AWX.
func (s *Semaphore) GetJobTemplates(ctx context.Context) (*Paginated, error) {
	var raw []map[string]any
	path := fmt.Sprintf("/project/%d/templates", s.projectID)
	ok, err := s.request(ctx, http.MethodGet, path, nil, &raw)
	if err != nil || !ok {
		return nil, err
	}
	inventories := s.inventoriesByID(ctx)
	projectName := s.projectName(ctx)
	results := make([]map[string]any, 0, len(raw))
	for _, t := range raw {
		results = append(results, s.mapTemplate(t, inventories, projectName))
	}
	return &Paginated{Count: len(results), Results: results}, nil
}

// GetJobTemplate retourne un template par id.
func (s *Semaphore) GetJobTemplate(ctx context.Context, templateID int) (map[string]any, error) {
	var raw map[string]any
	path := fmt.Sprintf("/project/%d/templates/%d", s.projectID, templateID)
	ok, err := s.request(ctx, http.MethodGet, path, nil, &raw)
	if err != nil || !ok {
		return nil, err
	}
	return s.mapTemplate(raw, s.inventoriesByID(ctx), s.projectName(ctx)), nil
}

// LaunchJobTemplate crée une tâche Semaphore depuis un template et retourne la
// tâche brute (l'appelant récupère ensuite le détail enrichi via GetJob).
func (s *Semaphore) LaunchJobTemplate(ctx context.Context, templateID int, extraVars map[string]any) (map[string]any, error) {
	payload := map[string]any{"template_id": templateID}
	if len(extraVars) > 0 {
		env, err := json.Marshal(extraVars)
		if err != nil {
			return nil, err
		}
		payload["environment"] = string(env)
	}
	var raw map[string]any
	path := fmt.Sprintf("/project/%d/tasks", s.projectID)
	ok, err := s.request(ctx, http.MethodPost, path, payload, &raw)
	if err != nil || !ok {
		return nil, err
	}
	return raw, nil
}

func (s *Semaphore) mapTemplate(t map[string]any, inventories map[int]map[string]any, projectName string) map[string]any {
	var inventoryName any
	inventoryID, hasInventory := t["inventory_id"]
	if hasInventory {
		if inv, found := inventories[intVal(inventoryID)]; found {
			inventoryName = inv["name"]
		}
	}

	name := str(t["alias"])
	if name == "" {
		name = str(t["name"])
	}

	projectID := intVal(t["project_id"])
	if projectID == 0 {
		projectID = s.projectID
	}

	summary := map[string]any{"inventory": nil, "project": nil}
	if inventoryName != nil {
		summary["inventory"] = map[string]any{"name": inventoryName}
	}
	if projectName != "" {
		summary["project"] = map[string]any{"name": projectName}
	}

	return map[string]any{
		"id":             t["id"],
		"name":           name,
		"description":    str(t["description"]),
		"playbook":       str(t["playbook"]),
		"repository_id":  t["repository_id"],
		"inventory":      inventoryID,
		"inventory_name": inventoryName,
		"project":        projectID,
		"project_name":   projectName,
		"summary_fields": summary,
	}
}

// ── Dépôts ───────────────────────────────────────────────────────────────────

// GetRepository retourne un dépôt Git configuré pour le projet Semaphore.
func (s *Semaphore) GetRepository(ctx context.Context, repositoryID int) (map[string]any, error) {
	var raw map[string]any
	path := fmt.Sprintf("/project/%d/repositories/%d", s.projectID, repositoryID)
	ok, err := s.request(ctx, http.MethodGet, path, nil, &raw)
	if err != nil || !ok {
		return nil, err
	}
	return raw, nil
}

// ── Inventaires ──────────────────────────────────────────────────────────────

// GetInventories liste les inventaires du projet au format AWX.
func (s *Semaphore) GetInventories(ctx context.Context) (*Paginated, error) {
	var raw []map[string]any
	path := fmt.Sprintf("/project/%d/inventory", s.projectID)
	ok, err := s.request(ctx, http.MethodGet, path, nil, &raw)
	if err != nil || !ok {
		return nil, err
	}
	results := make([]map[string]any, 0, len(raw))
	for _, i := range raw {
		results = append(results, mapInventory(i))
	}
	return &Paginated{Count: len(results), Results: results}, nil
}

// GetInventory retourne un inventaire par id.
func (s *Semaphore) GetInventory(ctx context.Context, inventoryID int) (map[string]any, error) {
	var raw map[string]any
	path := fmt.Sprintf("/project/%d/inventory/%d", s.projectID, inventoryID)
	ok, err := s.request(ctx, http.MethodGet, path, nil, &raw)
	if err != nil || !ok {
		return nil, err
	}
	return mapInventory(raw), nil
}

// GetInventoryHosts extrait les hôtes du contenu brut d'un inventaire statique.
func (s *Semaphore) GetInventoryHosts(ctx context.Context, inventoryID int) (*Paginated, error) {
	inventory, err := s.GetInventory(ctx, inventoryID)
	if err != nil {
		return nil, err
	}
	if inventory == nil {
		return EmptyPaginated(), nil
	}
	content := str(inventory["_raw_inventory"])
	hosts := []map[string]any{}
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "#") {
			continue
		}
		hosts = append(hosts, map[string]any{"name": line})
	}
	return &Paginated{Count: len(hosts), Results: hosts}, nil
}

func mapInventory(i map[string]any) map[string]any {
	invType := str(i["type"])
	if invType == "" {
		invType = "file"
	}
	return map[string]any{
		"id":             i["id"],
		"name":           str(i["name"]),
		"type":           invType,
		"_raw_inventory": str(i["inventory"]),
		"host_filter":    nil,
	}
}

// ── Stubs de compatibilité AWX ───────────────────────────────────────────────

// GetCredentials : Semaphore gère ses accès en interne, la liste AWX est vide.
func (s *Semaphore) GetCredentials(ctx context.Context) *Paginated {
	return EmptyPaginated()
}

// GetOrganizations : Semaphore n'a pas d'organisations, on expose "Default".
func (s *Semaphore) GetOrganizations(ctx context.Context) *Paginated {
	return &Paginated{
		Count:   1,
		Results: []map[string]any{{"id": 1, "name": "Default"}},
	}
}

// ── Helpers internes ─────────────────────────────────────────────────────────

func (s *Semaphore) templatesByID(ctx context.Context) map[int]map[string]any {
	var raw []map[string]any
	path := fmt.Sprintf("/project/%d/templates", s.projectID)
	if ok, err := s.request(ctx, http.MethodGet, path, nil, &raw); err != nil || !ok {
		return map[int]map[string]any{}
	}
	result := make(map[int]map[string]any, len(raw))
	for _, t := range raw {
		if id := intVal(t["id"]); id != 0 {
			result[id] = t
		}
	}
	return result
}

func (s *Semaphore) inventoriesByID(ctx context.Context) map[int]map[string]any {
	var raw []map[string]any
	path := fmt.Sprintf("/project/%d/inventory", s.projectID)
	if ok, err := s.request(ctx, http.MethodGet, path, nil, &raw); err != nil || !ok {
		return map[int]map[string]any{}
	}
	result := make(map[int]map[string]any, len(raw))
	for _, i := range raw {
		if id := intVal(i["id"]); id != 0 {
			result[id] = i
		}
	}
	return result
}

func (s *Semaphore) projectName(ctx context.Context) string {
	s.nameOnce.Do(func() {
		project, err := s.GetProject(ctx, s.projectID)
		if err == nil && project != nil {
			s.nameCached = str(project["name"])
		}
	})
	return s.nameCached
}

// computeElapsed calcule la durée en secondes entre deux horodatages ISO 8601.
func computeElapsed(started, finished string) float64 {
	if started == "" || finished == "" {
		return 0
	}
	start, err := time.Parse(time.RFC3339, started)
	if err != nil {
		return 0
	}
	end, err := time.Parse(time.RFC3339, finished)
	if err != nil {
		return 0
	}
	elapsed := end.Sub(start).Seconds()
	if elapsed < 0 {
		return 0
	}
	return elapsed
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func intVal(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i)
		}
	}
	return 0
}

func floatVal(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		if f, err := n.Float64(); err == nil {
			return f
		}
	}
	return 0
}
