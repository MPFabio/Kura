package service

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/modulops/k8s-service/internal/k8s"
	"github.com/modulops/k8s-service/internal/models"
)

// DiscoveryService agrège, pour le cluster actif du client, les Applications
// ArgoCD déployées et les composants d'observabilité reconnus parmi elles
// (Prometheus/VictoriaMetrics, Grafana, Loki, Tempo). Il ne nécessite aucune
// configuration côté client : la détection se fait par labels Kubernetes
// standards des charts Helm courants.
type DiscoveryService struct {
	clusterService *ClusterService
	argocdService  *ArgoCDService
}

// NewDiscoveryService crée un nouveau service de découverte.
func NewDiscoveryService(clusterService *ClusterService, argocdService *ArgoCDService) *DiscoveryService {
	return &DiscoveryService{
		clusterService: clusterService,
		argocdService:  argocdService,
	}
}

// observabilityTargets liste les rôles d'observabilité que Kura sait
// reconnaître automatiquement dans le cluster client.
func observabilityTargets() []k8s.ObservabilityTarget {
	return []k8s.ObservabilityTarget{
		k8s.PrometheusTarget,
		k8s.GrafanaTarget,
		k8s.LokiTarget,
		k8s.TempoTarget,
	}
}

// clientsetForActiveCluster construit un clientset Kubernetes pour le cluster actif.
func (s *DiscoveryService) clientsetForActiveCluster(ctx context.Context) (*kubernetes.Clientset, error) {
	cluster, err := s.clusterService.GetActiveCluster(ctx)
	if err != nil {
		return nil, fmt.Errorf("aucun cluster actif: %w", err)
	}

	kubeconfigContent, err := s.clusterService.GetPortableKubeconfig(ctx, cluster)
	if err != nil {
		return nil, fmt.Errorf("préparation du kubeconfig: %w", err)
	}

	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigContent))
	if err != nil {
		return nil, fmt.Errorf("chargement du kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("création du client Kubernetes: %w", err)
	}

	return clientset, nil
}

// GetReport retourne les Applications ArgoCD du cluster actif ainsi que les
// composants d'observabilité détectés automatiquement par labels.
func (s *DiscoveryService) GetReport(ctx context.Context) (*models.DiscoveryReport, error) {
	report := &models.DiscoveryReport{
		Applications:  []models.ArgoApplication{},
		Observability: []models.DiscoveredComponent{},
	}

	if apps, err := s.argocdService.ListApplications(ctx); err == nil {
		report.Applications = apps
	}

	clientset, err := s.clientsetForActiveCluster(ctx)
	if err != nil {
		return nil, err
	}

	for _, target := range observabilityTargets() {
		component := models.DiscoveredComponent{Name: target.Name}
		if pod, err := k8s.FindPod(ctx, clientset, target); err == nil {
			component.Found = true
			component.Namespace = pod.Namespace
			component.PodName = pod.Name
		}
		report.Observability = append(report.Observability, component)
	}

	return report, nil
}
