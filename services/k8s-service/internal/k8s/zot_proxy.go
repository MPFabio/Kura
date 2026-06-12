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

// ZotNamespace est le namespace dans lequel le registre Zot est déployé sur le cluster client.
const ZotNamespace = "zot"

// zotLabelSelector sélectionne le pod du registre Zot (label posé par le chart Helm officiel).
const zotLabelSelector = "app.kubernetes.io/name=zot"

// zotPort est le port HTTP exposé par Zot.
const zotPort = 5000

// ZotProxy gère un port-forward SPDY vers le pod Zot d'un cluster.
type ZotProxy struct {
	localPort int
	stopCh    chan struct{}
	readyCh   chan struct{}
	errCh     chan error
}

// NewZotPortForwarder crée un nouveau proxy de port-forward vers le pod Zot.
func NewZotPortForwarder(restConfig *rest.Config, clientset *kubernetes.Clientset) (*ZotProxy, error) {
	pods, err := clientset.CoreV1().Pods(ZotNamespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: zotLabelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("recherche du pod zot: %w", err)
	}
	var podName string
	for _, p := range pods.Items {
		if p.Status.Phase == "Running" {
			podName = p.Name
			break
		}
	}
	if podName == "" {
		return nil, fmt.Errorf("aucun pod zot en cours d'exécution dans le namespace %s", ZotNamespace)
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
		Path:   fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward", ZotNamespace, podName),
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", requestURL)

	proxy := &ZotProxy{
		stopCh:  make(chan struct{}),
		readyCh: make(chan struct{}),
		errCh:   make(chan error, 1),
	}

	pf, err := portforward.New(dialer, []string{fmt.Sprintf("0:%d", zotPort)}, proxy.stopCh, proxy.readyCh, io.Discard, io.Discard)
	if err != nil {
		return nil, fmt.Errorf("création du port-forward: %w", err)
	}

	go func() {
		proxy.errCh <- pf.ForwardPorts()
	}()

	select {
	case <-proxy.readyCh:
	case err := <-proxy.errCh:
		return nil, fmt.Errorf("échec du port-forward vers zot: %w", err)
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

// BaseURL retourne l'URL de base locale à utiliser pour appeler l'API Zot.
// Zot sert son API OCI Distribution en HTTP simple.
func (p *ZotProxy) BaseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", p.localPort)
}

// Stop arrête le port-forward.
func (p *ZotProxy) Stop() {
	close(p.stopCh)
}
