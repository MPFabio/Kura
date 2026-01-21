package k8s

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/modulops/k8s-service/internal/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// Client encapsule le client Kubernetes officiel.
type Client struct {
	clientset *kubernetes.Clientset
	restConfig *rest.Config
}

// NewClient crée un nouveau client Kubernetes à partir de la configuration.
func NewClient(cfg *config.Config) (*Client, error) {
	var (
		restConfig *rest.Config
		err        error
	)

	if cfg.InCluster {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("échec de la configuration in-cluster: %w", err)
		}
	} else {
		if cfg.KubeconfigPath == "" {
			return nil, fmt.Errorf("KUBECONFIG_PATH doit être défini lorsque K8S_INCLUSTER=false")
		}
		
		// Charger la configuration avec possibilité de désactiver TLS pour les clusters locaux
		configLoadingRules := &clientcmd.ClientConfigLoadingRules{
			ExplicitPath: cfg.KubeconfigPath,
		}
		configOverrides := &clientcmd.ConfigOverrides{}
		
		// Vérifier si le serveur utilise host.docker.internal ou 127.0.0.1
		rawConfig, err := configLoadingRules.Load()
		if err != nil {
			return nil, fmt.Errorf("échec du chargement du kubeconfig: %w", err)
		}
		
		// Déterminer le contexte à utiliser
		contextToUse := rawConfig.CurrentContext
		if contextToUse == "" && len(rawConfig.Contexts) > 0 {
			// Si aucun contexte actuel, utiliser le premier contexte disponible
			for name := range rawConfig.Contexts {
				contextToUse = name
				break
			}
		}
		
		if contextToUse != "" {
			currentContext := rawConfig.Contexts[contextToUse]
			if currentContext != nil {
				clusterName := currentContext.Cluster
				cluster := rawConfig.Clusters[clusterName]
				if cluster != nil && (strings.Contains(cluster.Server, "host.docker.internal") || strings.Contains(cluster.Server, "127.0.0.1")) {
					// Forcer InsecureSkipTLSVerify pour les clusters locaux
					configOverrides.ClusterInfo.InsecureSkipTLSVerify = true
					configOverrides.ClusterInfo.CertificateAuthorityData = nil
				}
				// S'assurer que le contexte est correctement défini
				configOverrides.CurrentContext = contextToUse
			}
		}
		
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(configLoadingRules, configOverrides)
		restConfig, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("échec du chargement du kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("échec de la création du clientset Kubernetes: %w", err)
	}

	return &Client{
		clientset:  clientset,
		restConfig: restConfig,
	}, nil
}

// ListNamespaces retourne la liste des namespaces.
func (c *Client) ListNamespaces(ctx context.Context) ([]corev1.Namespace, error) {
	nsList, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nsList.Items, nil
}

// ListPods retourne la liste des pods pour un namespace donné.
func (c *Client) ListPods(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	podList, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

// ListDeployments retourne la liste des deployments pour un namespace donné.
func (c *Client) ListDeployments(ctx context.Context, namespace string) ([]appsv1.Deployment, error) {
	deployList, err := c.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return deployList.Items, nil
}

// ListServices retourne la liste des services pour un namespace donné.
func (c *Client) ListServices(ctx context.Context, namespace string) ([]corev1.Service, error) {
	svcList, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return svcList.Items, nil
}

// ListConfigMaps retourne la liste des ConfigMaps pour un namespace donné.
func (c *Client) ListConfigMaps(ctx context.Context, namespace string) ([]corev1.ConfigMap, error) {
	cmList, err := c.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return cmList.Items, nil
}

// ListSecrets retourne la liste des secrets pour un namespace donné.
func (c *Client) ListSecrets(ctx context.Context, namespace string) ([]corev1.Secret, error) {
	secretList, err := c.clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return secretList.Items, nil
}

// ListNodes retourne la liste de tous les nodes du cluster.
func (c *Client) ListNodes(ctx context.Context) ([]corev1.Node, error) {
	nodeList, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}

// GetPod retourne les détails d'un pod spécifique.
func (c *Client) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return c.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetDeployment retourne les détails d'un deployment spécifique.
func (c *Client) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetService retourne les détails d'un service spécifique.
func (c *Client) GetService(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	return c.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetPodLogs retourne les logs d'un pod sous forme de stream.
func (c *Client) GetPodLogs(ctx context.Context, namespace, name, container string, tailLines *int64, follow bool) (io.ReadCloser, error) {
	opts := &corev1.PodLogOptions{
		Container: container,
		Follow:    follow,
	}
	if tailLines != nil {
		opts.TailLines = tailLines
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
	return req.Stream(ctx)
}

// GetPodLogsString retourne les logs d'un pod sous forme de string.
func (c *Client) GetPodLogsString(ctx context.Context, namespace, name, container string, tailLines *int64) (string, error) {
	opts := &corev1.PodLogOptions{
		Container: container,
	}
	if tailLines != nil {
		opts.TailLines = tailLines
	}

	req := c.clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetPodYAML retourne le YAML d'un pod.
func (c *Client) GetPodYAML(ctx context.Context, namespace, name string) (string, error) {
	pod, err := c.GetPod(ctx, namespace, name)
	if err != nil {
		return "", err
	}
	return toYAML(pod)
}

// GetDeploymentYAML retourne le YAML d'un deployment.
func (c *Client) GetDeploymentYAML(ctx context.Context, namespace, name string) (string, error) {
	deployment, err := c.GetDeployment(ctx, namespace, name)
	if err != nil {
		return "", err
	}
	return toYAML(deployment)
}

// GetServiceYAML retourne le YAML d'un service.
func (c *Client) GetServiceYAML(ctx context.Context, namespace, name string) (string, error) {
	service, err := c.GetService(ctx, namespace, name)
	if err != nil {
		return "", err
	}
	return toYAML(service)
}

// toYAML convertit un objet Kubernetes en YAML.
func toYAML(obj runtime.Object) (string, error) {
	// Créer un sérialiseur JSON
	serializer := json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme.Scheme,
		scheme.Scheme,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: false,
		},
	)

	// Sérialiser directement en YAML
	var buf strings.Builder
	err := serializer.Encode(obj, &buf)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la sérialisation YAML: %w", err)
	}

	return buf.String(), nil
}

// ScaleDeployment modifie le nombre de replicas d'un deployment.
func (c *Client) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	deployment, err := c.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("erreur lors de la récupération du deployment: %w", err)
	}

	deployment.Spec.Replicas = &replicas
	_, err = c.clientset.AppsV1().Deployments(namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du deployment: %w", err)
	}

	return nil
}

// DeletePod supprime un pod.
func (c *Client) DeletePod(ctx context.Context, namespace, name string) error {
	return c.clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// DeleteDeployment supprime un deployment.
func (c *Client) DeleteDeployment(ctx context.Context, namespace, name string) error {
	return c.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// DeleteService supprime un service.
func (c *Client) DeleteService(ctx context.Context, namespace, name string) error {
	return c.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// ListEvents retourne les événements pour un namespace donné.
func (c *Client) ListEvents(ctx context.Context, namespace string) ([]corev1.Event, error) {
	eventList, err := c.clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return eventList.Items, nil
}

// ExecPod exécute une commande dans un pod et retourne les streams stdin/stdout/stderr.
func (c *Client) ExecPod(ctx context.Context, namespace, name, container string, command []string, stdin io.Reader, stdout, stderr io.Writer, tty bool, sizeQueue remotecommand.TerminalSizeQueue) error {
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdin:     stdin != nil,
		Stdout:    stdout != nil,
		Stderr:    stderr != nil,
		TTY:       tty,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.restConfig, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("erreur lors de la création de l'exécuteur: %w", err)
	}

	return exec.Stream(remotecommand.StreamOptions{
		Stdin:             stdin,
		Stdout:            stdout,
		Stderr:            stderr,
		TerminalSizeQueue: sizeQueue,
		Tty:               tty,
	})
}

