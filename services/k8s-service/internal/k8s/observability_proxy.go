package k8s

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// ObservabilityTarget décrit un composant de la stack d'observabilité du
// cluster client (Prometheus/VictoriaMetrics, Loki ou Tempo) déployé via le
// catalogue Helm ArgoCD (ex: kube-prometheus-stack, victoria-metrics-single,
// grafana/loki, grafana/tempo). Le namespace n'étant pas figé (choisi par
// l'utilisateur à la création de l'Application ArgoCD), la recherche du pod
// se fait par label sur tous les namespaces. Plusieurs charts Helm pouvant
// fournir le même rôle avec des labels différents (ex: Prometheus vs
// VictoriaMetrics), plusieurs sélecteurs candidats sont essayés dans l'ordre.
type ObservabilityTarget struct {
	// Name identifie la cible (pour les messages d'erreur).
	Name string
	// LabelSelectors sélectionnent le pod du composant, essayés dans l'ordre
	// jusqu'à trouver un pod en cours d'exécution.
	LabelSelectors []string
	// Port est le port HTTP exposé par le pod.
	Port int
}

var (
	// PrometheusTarget cible le pod du moteur de métriques PromQL déployé dans
	// le cluster client : kube-prometheus-stack (Prometheus) ou un chart
	// VictoriaMetrics (vmsingle/victoria-metrics-single), tous deux exposant
	// une API compatible PromQL sur /api/v1/query.
	PrometheusTarget = ObservabilityTarget{
		Name: "Prometheus",
		LabelSelectors: []string{
			"app.kubernetes.io/name=prometheus",
			"app.kubernetes.io/name=vmsingle",
			"app.kubernetes.io/name=victoria-metrics-single",
			"app=vmsingle",
		},
		Port: 9090,
	}
	// LokiTarget cible le pod loki déployé par le chart grafana/loki.
	LokiTarget = ObservabilityTarget{
		Name:           "Loki",
		LabelSelectors: []string{"app.kubernetes.io/name=loki"},
		Port:           3100,
	}
	// TempoTarget cible le pod tempo déployé par le chart grafana/tempo.
	TempoTarget = ObservabilityTarget{
		Name:           "Tempo",
		LabelSelectors: []string{"app.kubernetes.io/name=tempo"},
		Port:           3200,
	}
	// GrafanaTarget cible le pod grafana déployé par kube-prometheus-stack ou
	// le chart grafana/grafana autonome.
	GrafanaTarget = ObservabilityTarget{
		Name:           "Grafana",
		LabelSelectors: []string{"app.kubernetes.io/name=grafana"},
		Port:           3000,
	}
)

// ObservabilityProxy gère un port-forward SPDY vers un pod d'observabilité
// (Prometheus, Loki ou Tempo) du cluster client.
type ObservabilityProxy struct {
	localPort int
	stopCh    chan struct{}
	readyCh   chan struct{}
	errCh     chan error
}

// FindPod recherche, dans tous les namespaces, le premier pod en cours
// d'exécution correspondant à l'un des sélecteurs candidats de la cible
// donnée (essayés dans l'ordre).
func FindPod(ctx context.Context, clientset *kubernetes.Clientset, target ObservabilityTarget) (*corev1.Pod, error) {
	for _, selector := range target.LabelSelectors {
		pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
		if err != nil {
			return nil, fmt.Errorf("recherche du pod %s: %w", target.Name, err)
		}

		for i := range pods.Items {
			if pods.Items[i].Status.Phase == corev1.PodRunning {
				return &pods.Items[i], nil
			}
		}
	}
	return nil, fmt.Errorf("aucun pod %s en cours d'exécution dans le cluster (stack d'observabilité du projet non déployée ?)", target.Name)
}

// NewObservabilityPortForwarder recherche, dans tous les namespaces, le
// premier pod en cours d'exécution correspondant à la cible donnée, et ouvre
// un port-forward vers son port.
func NewObservabilityPortForwarder(restConfig *rest.Config, clientset *kubernetes.Clientset, target ObservabilityTarget) (*ObservabilityProxy, error) {
	pod, err := FindPod(context.Background(), clientset, target)
	if err != nil {
		return nil, err
	}

	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return nil, fmt.Errorf("création du round-tripper SPDY: %w", err)
	}

	parsedURL, err := url.Parse(restConfig.Host)
	if err != nil {
		return nil, fmt.Errorf("analyse de l'URL du serveur API: %w", err)
	}

	requestURL := &url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
		Path:   fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", pod.Namespace, pod.Name),
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", requestURL)

	proxy := &ObservabilityProxy{
		stopCh:  make(chan struct{}),
		readyCh: make(chan struct{}),
		errCh:   make(chan error, 1),
	}

	pf, err := portforward.New(dialer, []string{fmt.Sprintf("0:%d", target.Port)}, proxy.stopCh, proxy.readyCh, io.Discard, io.Discard)
	if err != nil {
		return nil, fmt.Errorf("création du port-forward: %w", err)
	}

	go func() {
		proxy.errCh <- pf.ForwardPorts()
	}()

	select {
	case <-proxy.readyCh:
	case err := <-proxy.errCh:
		return nil, fmt.Errorf("échec du port-forward vers %s: %w", target.Name, err)
	}

	ports, err := pf.GetPorts()
	if err != nil {
		proxy.Stop()
		return nil, fmt.Errorf("lecture du port local alloué: %w", err)
	}
	if len(ports) == 0 {
		proxy.Stop()
		return nil, fmt.Errorf("aucun port local alloué pour le port-forward")
	}

	proxy.localPort = int(ports[0].Local)

	return proxy, nil
}

// BaseURL retourne l'URL de base locale à utiliser pour appeler le composant.
func (p *ObservabilityProxy) BaseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", p.localPort)
}

// Stop arrête le port-forward.
func (p *ObservabilityProxy) Stop() {
	close(p.stopCh)
}
