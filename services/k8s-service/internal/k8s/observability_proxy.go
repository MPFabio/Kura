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
// cluster client (Prometheus, Loki ou Tempo) déployé via le catalogue Helm
// ArgoCD (ex: kube-prometheus-stack, grafana/loki, grafana/tempo). Le
// namespace n'étant pas figé (choisi par l'utilisateur à la création de
// l'Application ArgoCD), la recherche du pod se fait par label sur tous les
// namespaces.
type ObservabilityTarget struct {
	// Name identifie la cible (pour les messages d'erreur).
	Name string
	// LabelSelector sélectionne le pod du composant.
	LabelSelector string
	// Port est le port HTTP exposé par le pod.
	Port int
}

var (
	// PrometheusTarget cible le pod prometheus déployé par kube-prometheus-stack.
	PrometheusTarget = ObservabilityTarget{
		Name:          "Prometheus",
		LabelSelector: "app.kubernetes.io/name=prometheus",
		Port:          9090,
	}
	// LokiTarget cible le pod loki déployé par le chart grafana/loki.
	LokiTarget = ObservabilityTarget{
		Name:          "Loki",
		LabelSelector: "app.kubernetes.io/name=loki",
		Port:          3100,
	}
	// TempoTarget cible le pod tempo déployé par le chart grafana/tempo.
	TempoTarget = ObservabilityTarget{
		Name:          "Tempo",
		LabelSelector: "app.kubernetes.io/name=tempo",
		Port:          3200,
	}
	// GrafanaTarget cible le pod grafana déployé par kube-prometheus-stack.
	GrafanaTarget = ObservabilityTarget{
		Name:          "Grafana",
		LabelSelector: "app.kubernetes.io/name=grafana",
		Port:          3000,
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

// NewObservabilityPortForwarder recherche, dans tous les namespaces, le
// premier pod en cours d'exécution correspondant à la cible donnée, et ouvre
// un port-forward vers son port.
func NewObservabilityPortForwarder(restConfig *rest.Config, clientset *kubernetes.Clientset, target ObservabilityTarget) (*ObservabilityProxy, error) {
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
		LabelSelector: target.LabelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("recherche du pod %s: %w", target.Name, err)
	}

	var pod *corev1.Pod
	for i := range pods.Items {
		if pods.Items[i].Status.Phase == corev1.PodRunning {
			pod = &pods.Items[i]
			break
		}
	}
	if pod == nil {
		return nil, fmt.Errorf("aucun pod %s en cours d'exécution dans le cluster (stack d'observabilité du projet non déployée ?)", target.Name)
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
