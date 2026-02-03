package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/modulops/k8s-service/internal/config"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/remotecommand"
)

// K8sClient définit les méthodes du client Kubernetes utilisées par le service.
type K8sClient interface {
	ListNamespaces(ctx context.Context) ([]corev1.Namespace, error)
	ListPods(ctx context.Context, namespace string) ([]corev1.Pod, error)
	ListDeployments(ctx context.Context, namespace string) ([]appsv1.Deployment, error)
	ListServices(ctx context.Context, namespace string) ([]corev1.Service, error)
	ListConfigMaps(ctx context.Context, namespace string) ([]corev1.ConfigMap, error)
	ListSecrets(ctx context.Context, namespace string) ([]corev1.Secret, error)
	ListNodes(ctx context.Context) ([]corev1.Node, error)
	GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error)
	GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error)
	GetService(ctx context.Context, namespace, name string) (*corev1.Service, error)
	GetPodLogs(ctx context.Context, namespace, name, container string, tailLines *int64, follow bool) (io.ReadCloser, error)
	GetPodLogsString(ctx context.Context, namespace, name, container string, tailLines *int64) (string, error)
	GetPodYAML(ctx context.Context, namespace, name string) (string, error)
	GetDeploymentYAML(ctx context.Context, namespace, name string) (string, error)
	GetServiceYAML(ctx context.Context, namespace, name string) (string, error)
	GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error)
	GetSecret(ctx context.Context, namespace, name string) (*corev1.Secret, error)
	GetNode(ctx context.Context, name string) (*corev1.Node, error)
	GetConfigMapYAML(ctx context.Context, namespace, name string) (string, error)
	GetSecretYAML(ctx context.Context, namespace, name string) (string, error)
	GetNodeYAML(ctx context.Context, name string) (string, error)
	ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error
	DeletePod(ctx context.Context, namespace, name string) error
	DeleteDeployment(ctx context.Context, namespace, name string) error
	DeleteService(ctx context.Context, namespace, name string) error
	ListEvents(ctx context.Context, namespace string) ([]corev1.Event, error)
	ExecPod(ctx context.Context, namespace, name, container string, command []string, stdin io.Reader, stdout, stderr io.Writer, tty bool, sizeQueue remotecommand.TerminalSizeQueue) error
}

// Cache définit l'interface minimale du cache utilisée par le service.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

// K8sService contient la logique métier autour de Kubernetes.
type K8sService struct {
	k8sClient K8sClient
	cache     Cache
	cfg       *config.Config
}

// NewK8sService crée un nouveau service Kubernetes.
func NewK8sService(k8sClient K8sClient, redisClient Cache, cfg *config.Config) *K8sService {
	return &K8sService{
		k8sClient: k8sClient,
		cache:     redisClient,
		cfg:       cfg,
	}
}

// NamespaceDTO est un DTO simplifié pour exposer les namespaces.
type NamespaceDTO struct {
	Name              string            `json:"name"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels,omitempty"`
}

// PodDTO est un DTO simplifié pour exposer les pods.
type PodDTO struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	NodeName          string            `json:"nodeName,omitempty"`
	Phase             string            `json:"phase"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels,omitempty"`
}

// ListNamespaces retourne la liste des namespaces, avec cache Redis.
func (s *K8sService) ListNamespaces(ctx context.Context) ([]NamespaceDTO, error) {
	cacheKey := "k8s:namespaces"

	// Tentative de lecture depuis le cache
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var items []NamespaceDTO
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	// Fallback sur Kubernetes
	namespaces, err := s.k8sClient.ListNamespaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des namespaces: %w", err)
	}

	result := make([]NamespaceDTO, 0, len(namespaces))
	for _, ns := range namespaces {
		result = append(result, NamespaceDTO{
			Name:              ns.Name,
			CreationTimestamp: ns.CreationTimestamp.Time,
			Labels:            ns.Labels,
		})
	}

	// Stocker en cache (best-effort)
	if data, err := json.Marshal(result); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), s.cfg.CacheTTL)
	}

	return result, nil
}

// ListPods retourne la liste des pods pour un namespace donné, avec cache Redis.
func (s *K8sService) ListPods(ctx context.Context, namespace string) ([]PodDTO, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace requis")
	}

	cacheKey := fmt.Sprintf("k8s:pods:%s", namespace)

	// Tentative de lecture depuis le cache
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var items []PodDTO
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	// Fallback sur Kubernetes
	pods, err := s.k8sClient.ListPods(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des pods: %w", err)
	}

	result := make([]PodDTO, 0, len(pods))
	for _, pod := range pods {
		result = append(result, PodDTO{
			Name:              pod.Name,
			Namespace:         pod.Namespace,
			NodeName:          pod.Spec.NodeName,
			Phase:             string(pod.Status.Phase),
			CreationTimestamp: pod.CreationTimestamp.Time,
			Labels:            pod.Labels,
		})
	}

	// Stocker en cache (best-effort)
	if data, err := json.Marshal(result); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), s.cfg.CacheTTL)
	}

	return result, nil
}

// DeploymentDTO est un DTO simplifié pour exposer les deployments.
type DeploymentDTO struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"readyReplicas"`
	AvailableReplicas int32             `json:"availableReplicas"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels,omitempty"`
}

// ServiceDTO est un DTO simplifié pour exposer les services.
type ServiceDTO struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Type              string            `json:"type"`
	ClusterIP         string            `json:"clusterIP"`
	Ports             []ServicePortDTO  `json:"ports,omitempty"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels,omitempty"`
}

// ServicePortDTO représente un port de service.
type ServicePortDTO struct {
	Name       string `json:"name,omitempty"`
	Port       int32  `json:"port"`
	Protocol   string `json:"protocol"`
	TargetPort string `json:"targetPort,omitempty"`
}

// ConfigMapDTO est un DTO simplifié pour exposer les ConfigMaps.
type ConfigMapDTO struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	DataKeys          []string          `json:"dataKeys"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels,omitempty"`
}

// SecretDTO est un DTO simplifié pour exposer les secrets.
type SecretDTO struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Type              string            `json:"type"`
	DataKeys          []string          `json:"dataKeys"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels,omitempty"`
}

// NodeDTO est un DTO simplifié pour exposer les nodes.
type NodeDTO struct {
	Name              string            `json:"name"`
	Status            string            `json:"status"`
	KubeletVersion    string            `json:"kubeletVersion"`
	OSImage           string            `json:"osImage"`
	Architecture      string            `json:"architecture"`
	CPU               string            `json:"cpu"`
	Memory            string            `json:"memory"`
	Pods              string            `json:"pods"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels,omitempty"`
}

// ListDeployments retourne la liste des deployments pour un namespace donné.
func (s *K8sService) ListDeployments(ctx context.Context, namespace string) ([]DeploymentDTO, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace requis")
	}

	cacheKey := fmt.Sprintf("k8s:deployments:%s", namespace)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var items []DeploymentDTO
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	deployments, err := s.k8sClient.ListDeployments(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des deployments: %w", err)
	}

	result := make([]DeploymentDTO, 0, len(deployments))
	for _, dep := range deployments {
		replicas := int32(0)
		if dep.Spec.Replicas != nil {
			replicas = *dep.Spec.Replicas
		}
		result = append(result, DeploymentDTO{
			Name:              dep.Name,
			Namespace:         dep.Namespace,
			Replicas:          replicas,
			ReadyReplicas:     dep.Status.ReadyReplicas,
			AvailableReplicas: dep.Status.AvailableReplicas,
			CreationTimestamp: dep.CreationTimestamp.Time,
			Labels:            dep.Labels,
		})
	}

	if data, err := json.Marshal(result); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), s.cfg.CacheTTL)
	}

	return result, nil
}

// ListServices retourne la liste des services pour un namespace donné.
func (s *K8sService) ListServices(ctx context.Context, namespace string) ([]ServiceDTO, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace requis")
	}

	cacheKey := fmt.Sprintf("k8s:services:%s", namespace)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var items []ServiceDTO
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	services, err := s.k8sClient.ListServices(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des services: %w", err)
	}

	result := make([]ServiceDTO, 0, len(services))
	for _, svc := range services {
		ports := make([]ServicePortDTO, 0, len(svc.Spec.Ports))
		for _, p := range svc.Spec.Ports {
			targetPort := ""
			if p.TargetPort.Type == intstr.Int {
				targetPort = fmt.Sprintf("%d", p.TargetPort.IntVal)
			} else {
				targetPort = p.TargetPort.StrVal
			}
			ports = append(ports, ServicePortDTO{
				Name:       p.Name,
				Port:       p.Port,
				Protocol:   string(p.Protocol),
				TargetPort: targetPort,
			})
		}
		result = append(result, ServiceDTO{
			Name:              svc.Name,
			Namespace:         svc.Namespace,
			Type:              string(svc.Spec.Type),
			ClusterIP:         svc.Spec.ClusterIP,
			Ports:             ports,
			CreationTimestamp: svc.CreationTimestamp.Time,
			Labels:            svc.Labels,
		})
	}

	if data, err := json.Marshal(result); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), s.cfg.CacheTTL)
	}

	return result, nil
}

// ListConfigMaps retourne la liste des ConfigMaps pour un namespace donné.
func (s *K8sService) ListConfigMaps(ctx context.Context, namespace string) ([]ConfigMapDTO, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace requis")
	}

	cacheKey := fmt.Sprintf("k8s:configmaps:%s", namespace)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var items []ConfigMapDTO
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	configMaps, err := s.k8sClient.ListConfigMaps(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des ConfigMaps: %w", err)
	}

	result := make([]ConfigMapDTO, 0, len(configMaps))
	for _, cm := range configMaps {
		keys := make([]string, 0, len(cm.Data))
		for k := range cm.Data {
			keys = append(keys, k)
		}
		result = append(result, ConfigMapDTO{
			Name:              cm.Name,
			Namespace:         cm.Namespace,
			DataKeys:          keys,
			CreationTimestamp: cm.CreationTimestamp.Time,
			Labels:            cm.Labels,
		})
	}

	if data, err := json.Marshal(result); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), s.cfg.CacheTTL)
	}

	return result, nil
}

// ListSecrets retourne la liste des secrets pour un namespace donné.
func (s *K8sService) ListSecrets(ctx context.Context, namespace string) ([]SecretDTO, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace requis")
	}

	cacheKey := fmt.Sprintf("k8s:secrets:%s", namespace)
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var items []SecretDTO
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	secrets, err := s.k8sClient.ListSecrets(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des secrets: %w", err)
	}

	result := make([]SecretDTO, 0, len(secrets))
	for _, secret := range secrets {
		keys := make([]string, 0, len(secret.Data))
		for k := range secret.Data {
			keys = append(keys, k)
		}
		result = append(result, SecretDTO{
			Name:              secret.Name,
			Namespace:         secret.Namespace,
			Type:              string(secret.Type),
			DataKeys:          keys,
			CreationTimestamp: secret.CreationTimestamp.Time,
			Labels:            secret.Labels,
		})
	}

	if data, err := json.Marshal(result); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), s.cfg.CacheTTL)
	}

	return result, nil
}

// ListNodes retourne la liste de tous les nodes du cluster.
func (s *K8sService) ListNodes(ctx context.Context) ([]NodeDTO, error) {
	cacheKey := "k8s:nodes"
	if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var items []NodeDTO
		if err := json.Unmarshal([]byte(cached), &items); err == nil {
			return items, nil
		}
	}

	nodes, err := s.k8sClient.ListNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des nodes: %w", err)
	}

	result := make([]NodeDTO, 0, len(nodes))
	for _, node := range nodes {
		status := "Unknown"
		for _, cond := range node.Status.Conditions {
			if cond.Type == corev1.NodeReady {
				if cond.Status == corev1.ConditionTrue {
					status = "Ready"
				} else {
					status = "NotReady"
				}
				break
			}
		}

		cpu := "0"
		memory := "0"
		pods := "0"
		for resName, res := range node.Status.Capacity {
			switch resName {
			case corev1.ResourceCPU:
				cpu = res.String()
			case corev1.ResourceMemory:
				memory = res.String()
			case corev1.ResourcePods:
				pods = res.String()
			}
		}

		kubeletVersion := node.Status.NodeInfo.KubeletVersion
		osImage := node.Status.NodeInfo.OSImage
		arch := node.Status.NodeInfo.Architecture

		result = append(result, NodeDTO{
			Name:              node.Name,
			Status:            status,
			KubeletVersion:    kubeletVersion,
			OSImage:           osImage,
			Architecture:      arch,
			CPU:               cpu,
			Memory:            memory,
			Pods:              pods,
			CreationTimestamp: node.CreationTimestamp.Time,
			Labels:            node.Labels,
		})
	}

	if data, err := json.Marshal(result); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(data), s.cfg.CacheTTL)
	}

	return result, nil
}

// GetPodLogs retourne les logs d'un pod.
func (s *K8sService) GetPodLogs(ctx context.Context, namespace, name, container string, tailLines *int64) (string, error) {
	return s.k8sClient.GetPodLogsString(ctx, namespace, name, container, tailLines)
}

// GetPodYAML retourne le YAML d'un pod.
func (s *K8sService) GetPodYAML(ctx context.Context, namespace, name string) (string, error) {
	return s.k8sClient.GetPodYAML(ctx, namespace, name)
}

// GetDeploymentYAML retourne le YAML d'un deployment.
func (s *K8sService) GetDeploymentYAML(ctx context.Context, namespace, name string) (string, error) {
	return s.k8sClient.GetDeploymentYAML(ctx, namespace, name)
}

// GetServiceYAML retourne le YAML d'un service.
func (s *K8sService) GetServiceYAML(ctx context.Context, namespace, name string) (string, error) {
	return s.k8sClient.GetServiceYAML(ctx, namespace, name)
}

// GetConfigMapYAML retourne le YAML d'un ConfigMap.
func (s *K8sService) GetConfigMapYAML(ctx context.Context, namespace, name string) (string, error) {
	return s.k8sClient.GetConfigMapYAML(ctx, namespace, name)
}

// GetSecretYAML retourne le YAML d'un Secret.
func (s *K8sService) GetSecretYAML(ctx context.Context, namespace, name string) (string, error) {
	return s.k8sClient.GetSecretYAML(ctx, namespace, name)
}

// GetNodeYAML retourne le YAML d'un Node.
func (s *K8sService) GetNodeYAML(ctx context.Context, name string) (string, error) {
	return s.k8sClient.GetNodeYAML(ctx, name)
}

// ScaleDeployment modifie le nombre de replicas d'un deployment.
func (s *K8sService) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
	return s.k8sClient.ScaleDeployment(ctx, namespace, name, replicas)
}

// DeletePod supprime un pod.
func (s *K8sService) DeletePod(ctx context.Context, namespace, name string) error {
	return s.k8sClient.DeletePod(ctx, namespace, name)
}

// DeleteDeployment supprime un deployment.
func (s *K8sService) DeleteDeployment(ctx context.Context, namespace, name string) error {
	return s.k8sClient.DeleteDeployment(ctx, namespace, name)
}

// DeleteService supprime un service.
func (s *K8sService) DeleteService(ctx context.Context, namespace, name string) error {
	return s.k8sClient.DeleteService(ctx, namespace, name)
}

// EventDTO est un DTO simplifié pour exposer les événements.
type EventDTO struct {
	Name               string    `json:"name"`
	Namespace          string    `json:"namespace"`
	Type               string    `json:"type"`
	Reason             string    `json:"reason"`
	Message            string    `json:"message"`
	FirstTimestamp     time.Time `json:"firstTimestamp"`
	LastTimestamp      time.Time `json:"lastTimestamp"`
	Count              int32     `json:"count"`
	InvolvedObject     string    `json:"involvedObject"`
	InvolvedObjectKind string    `json:"involvedObjectKind"`
}

// ListEvents retourne la liste des événements pour un namespace donné.
func (s *K8sService) ListEvents(ctx context.Context, namespace string) ([]EventDTO, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace requis")
	}

	events, err := s.k8sClient.ListEvents(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des événements: %w", err)
	}

	result := make([]EventDTO, 0, len(events))
	for _, event := range events {
		involvedObject := ""
		if event.InvolvedObject.Name != "" {
			involvedObject = fmt.Sprintf("%s/%s", event.InvolvedObject.Kind, event.InvolvedObject.Name)
		}
		result = append(result, EventDTO{
			Name:               event.Name,
			Namespace:          event.Namespace,
			Type:               event.Type,
			Reason:             event.Reason,
			Message:            event.Message,
			FirstTimestamp:     event.FirstTimestamp.Time,
			LastTimestamp:      event.LastTimestamp.Time,
			Count:              event.Count,
			InvolvedObject:     involvedObject,
			InvolvedObjectKind: event.InvolvedObject.Kind,
		})
	}

	return result, nil
}

// ExecPod exécute une commande dans un pod.
func (s *K8sService) ExecPod(ctx context.Context, namespace, name, container string, command []string, stdin io.Reader, stdout, stderr io.Writer, tty bool, sizeQueue remotecommand.TerminalSizeQueue) error {
	return s.k8sClient.ExecPod(ctx, namespace, name, container, command, stdin, stdout, stderr, tty, sizeQueue)
}

// GetPod retourne les détails d'un pod.
func (s *K8sService) GetPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return s.k8sClient.GetPod(ctx, namespace, name)
}

// GetDeployment retourne les détails d'un deployment.
func (s *K8sService) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return s.k8sClient.GetDeployment(ctx, namespace, name)
}
