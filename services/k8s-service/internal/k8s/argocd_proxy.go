package k8s

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// argoCDServerLabelSelector sélectionne le pod du serveur API ArgoCD.
const argoCDServerLabelSelector = "app.kubernetes.io/name=argocd-server"

// argoCDServerPort est le port HTTP/gRPC exposé par argocd-server.
const argoCDServerPort = 8080

// ArgoCDProxy gère un port-forward SPDY vers le pod argocd-server d'un cluster.
type ArgoCDProxy struct {
	localPort int
	stopCh    chan struct{}
	readyCh   chan struct{}
	errCh     chan error
}

// NewArgoCDPortForwarder crée un nouveau proxy de port-forward vers argocd-server.
func NewArgoCDPortForwarder(restConfig *rest.Config, clientset *kubernetes.Clientset) (*ArgoCDProxy, error) {
	pods, err := clientset.CoreV1().Pods(ArgoCDNamespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: argoCDServerLabelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("recherche du pod argocd-server: %w", err)
	}
	var podName string
	for _, p := range pods.Items {
		if p.Status.Phase == "Running" {
			podName = p.Name
			break
		}
	}
	if podName == "" {
		return nil, fmt.Errorf("aucun pod argocd-server en cours d'exécution dans le namespace %s", ArgoCDNamespace)
	}

	transport, upgrader, err := spdy.RoundTripperFor(restConfig)
	if err != nil {
		return nil, fmt.Errorf("création du round-tripper SPDY: %w", err)
	}

	hostIP := restConfig.Host
	parsedURL, err := url.Parse(hostIP)
	if err != nil {
		return nil, fmt.Errorf("analyse de l'URL du serveur API: %w", err)
	}

	requestURL := &url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
		Path:   fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", ArgoCDNamespace, podName),
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", requestURL)

	proxy := &ArgoCDProxy{
		stopCh:  make(chan struct{}),
		readyCh: make(chan struct{}),
		errCh:   make(chan error, 1),
	}

	pf, err := portforward.New(dialer, []string{fmt.Sprintf("0:%d", argoCDServerPort)}, proxy.stopCh, proxy.readyCh, io.Discard, io.Discard)
	if err != nil {
		return nil, fmt.Errorf("création du port-forward: %w", err)
	}

	go func() {
		proxy.errCh <- pf.ForwardPorts()
	}()

	select {
	case <-proxy.readyCh:
	case err := <-proxy.errCh:
		return nil, fmt.Errorf("échec du port-forward vers argocd-server: %w", err)
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

// BaseURL retourne l'URL de base locale à utiliser pour appeler l'API argocd-server.
// argocd-server sert son API en HTTPS (certificat auto-signé) même en local par défaut.
func (p *ArgoCDProxy) BaseURL() string {
	return fmt.Sprintf("https://127.0.0.1:%d", p.localPort)
}

// Stop arrête le port-forward.
func (p *ArgoCDProxy) Stop() {
	close(p.stopCh)
}
