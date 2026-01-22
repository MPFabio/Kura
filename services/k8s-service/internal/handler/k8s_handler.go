package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/modulops/k8s-service/internal/config"
	"github.com/modulops/k8s-service/internal/k8s"
	"github.com/modulops/k8s-service/internal/service"
)

// K8sHandler gère les requêtes HTTP liées à Kubernetes.
type K8sHandler struct {
	svc            *service.K8sService
	clusterService *service.ClusterService
	cfg            *config.Config
	redisClient    service.Cache
	mu             sync.RWMutex // Pour la sécurité des threads lors de la mise à jour de svc
}

// NewK8sHandler crée un nouveau handler Kubernetes.
func NewK8sHandler(svc *service.K8sService, clusterService *service.ClusterService, redisClient service.Cache, cfg *config.Config) *K8sHandler {
	return &K8sHandler{
		svc:            svc,
		clusterService: clusterService,
		cfg:            cfg,
		redisClient:    redisClient,
	}
}

// checkService vérifie si le service est disponible et le crée dynamiquement si nécessaire.
func (h *K8sHandler) checkService(c *gin.Context) bool {
	h.mu.RLock()
	svc := h.svc
	h.mu.RUnlock()

	// Si le service existe déjà, on l'utilise
	if svc != nil {
		return true
	}

	// Sinon, essayer de créer le service à partir du cluster actif
	if h.clusterService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Aucun cluster Kubernetes configuré. Veuillez ajouter un cluster d'abord.",
		})
		return false
	}

	ctx := c.Request.Context()
	activeCluster, err := h.clusterService.GetActiveCluster(ctx)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Aucun cluster Kubernetes actif. Veuillez activer un cluster d'abord.",
		})
		return false
	}

	// Créer un client Kubernetes temporaire à partir du cluster actif
	kubeconfigPath, err := h.clusterService.SaveKubeconfigToFile(ctx, activeCluster.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Erreur lors de la préparation du kubeconfig: %v", err),
		})
		return false
	}

	// Créer une config temporaire avec le kubeconfig
	tempCfg := &config.Config{
		KubeconfigPath: kubeconfigPath,
		InCluster:      false,
		RedisAddr:      h.cfg.RedisAddr,
		RedisPassword:  h.cfg.RedisPassword,
		RedisDB:        h.cfg.RedisDB,
		CacheTTL:       h.cfg.CacheTTL,
		ServerPort:     h.cfg.ServerPort,
		Environment:    h.cfg.Environment,
		LogLevel:       h.cfg.LogLevel,
	}

	k8sClient, err := k8s.NewClient(tempCfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Erreur lors de la connexion au cluster: %v", err),
		})
		return false
	}

	// Créer le service avec le nouveau client
	var k8sClientInterface service.K8sClient = k8sClient
	// h.redisClient est de type service.Cache, qui est compatible avec NewK8sService
	newSvc := service.NewK8sService(k8sClientInterface, h.redisClient, h.cfg)

	// Mettre à jour le service de manière thread-safe
	h.mu.Lock()
	h.svc = newSvc
	h.mu.Unlock()

	return true
}

// NamespaceResponse est le format de réponse attendu par le frontend.
type NamespaceResponse struct {
	Name      string `json:"name"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// GetNamespaces retourne la liste des namespaces.
func (h *K8sHandler) GetNamespaces(c *gin.Context) {
	if !h.checkService(c) {
		return
	}

	ctx := c.Request.Context()

	namespaces, err := h.svc.ListNamespaces(ctx)
	if err != nil {
		log.Printf("Erreur GetNamespaces: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "impossible de récupérer les namespaces"})
		return
	}

	// Mapper au format attendu par le frontend
	items := make([]NamespaceResponse, 0, len(namespaces))
	for _, ns := range namespaces {
		items = append(items, NamespaceResponse{
			Name:      ns.Name,
			Status:    "Active", // Par défaut, on considère que le namespace est actif
			CreatedAt: ns.CreationTimestamp.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}

// PodResponse est le format de réponse attendu par le frontend.
type PodResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Node      string `json:"node,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// GetPods retourne la liste des pods pour un namespace donné.
func (h *K8sHandler) GetPods(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	ctx := c.Request.Context()
	namespace := c.Param("namespace")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace requis"})
		return
	}

	pods, err := h.svc.ListPods(ctx, namespace)
	if err != nil {
		log.Printf("Erreur GetPods pour namespace %s: %v", namespace, err)
		
		// Vérifier si c'est une erreur de namespace non trouvé
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "notfound") {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("namespace '%s' non trouvé", namespace)})
			return
		}
		
		// Autres erreurs
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Mapper au format attendu par le frontend
	items := make([]PodResponse, 0, len(pods))
	for _, pod := range pods {
		items = append(items, PodResponse{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    pod.Phase, // Phase correspond au statut du pod
			Node:      pod.NodeName,
			CreatedAt: pod.CreationTimestamp.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
	})
}

// ReceiveEventWebhook est un endpoint webhook générique pour recevoir des événements Kubernetes.
// Il pourra être branché sur des Event Sinks, des contrôleurs ou d'autres producteurs d'événements.
func (h *K8sHandler) ReceiveEventWebhook(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload JSON invalide"})
		return
	}

	// Pour l'instant, on se contente de logguer l'événement.
	log.Printf("Webhook événement Kubernetes reçu: %+v", payload)

	c.JSON(http.StatusOK, gin.H{
		"status":  "received",
		"message": "événement enregistré",
	})
}

// GetDeployments retourne la liste des deployments pour un namespace donné.
func (h *K8sHandler) GetDeployments(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	ctx := c.Request.Context()
	namespace := c.Param("namespace")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace requis"})
		return
	}

	deployments, err := h.svc.ListDeployments(ctx, namespace)
	if err != nil {
		log.Printf("Erreur GetDeployments pour namespace %s: %v", namespace, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": deployments,
	})
}

// GetServices retourne la liste des services pour un namespace donné.
func (h *K8sHandler) GetServices(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	ctx := c.Request.Context()
	namespace := c.Param("namespace")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace requis"})
		return
	}

	services, err := h.svc.ListServices(ctx, namespace)
	if err != nil {
		log.Printf("Erreur GetServices pour namespace %s: %v", namespace, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": services,
	})
}

// GetConfigMaps retourne la liste des ConfigMaps pour un namespace donné.
func (h *K8sHandler) GetConfigMaps(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace requis"})
		return
	}

	configMaps, err := h.svc.ListConfigMaps(ctx, namespace)
	if err != nil {
		log.Printf("Erreur GetConfigMaps pour namespace %s: %v", namespace, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": configMaps,
	})
}

// GetSecrets retourne la liste des secrets pour un namespace donné.
func (h *K8sHandler) GetSecrets(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	ctx := c.Request.Context()
	namespace := c.Param("namespace")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace requis"})
		return
	}

	secrets, err := h.svc.ListSecrets(ctx, namespace)
	if err != nil {
		log.Printf("Erreur GetSecrets pour namespace %s: %v", namespace, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": secrets,
	})
}

// GetNodes retourne la liste de tous les nodes du cluster.
func (h *K8sHandler) GetNodes(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	ctx := c.Request.Context()

	nodes, err := h.svc.ListNodes(ctx)
	if err != nil {
		log.Printf("Erreur GetNodes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": nodes,
	})
}

// GetPodLogs retourne les logs d'un pod.
func (h *K8sHandler) GetPodLogs(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	name := c.Param("name")
	container := c.Query("container")
	
	// Paramètre optionnel pour limiter le nombre de lignes
	var tailLines *int64
	if tailStr := c.Query("tail"); tailStr != "" {
		if tail, err := strconv.ParseInt(tailStr, 10, 64); err == nil {
			tailLines = &tail
		}
	}

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace et name requis"})
		return
	}

	logs, err := h.svc.GetPodLogs(ctx, namespace, name, container, tailLines)
	if err != nil {
		log.Printf("Erreur GetPodLogs pour pod %s/%s: %v", namespace, name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
	})
}

// GetPodYAML retourne le YAML d'un pod.
func (h *K8sHandler) GetPodYAML(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace et name requis"})
		return
	}

	yaml, err := h.svc.GetPodYAML(ctx, namespace, name)
	if err != nil {
		log.Printf("Erreur GetPodYAML pour pod %s/%s: %v", namespace, name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/x-yaml", []byte(yaml))
}

// GetDeploymentYAML retourne le YAML d'un deployment.
func (h *K8sHandler) GetDeploymentYAML(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace et name requis"})
		return
	}

	yaml, err := h.svc.GetDeploymentYAML(ctx, namespace, name)
	if err != nil {
		log.Printf("Erreur GetDeploymentYAML pour deployment %s/%s: %v", namespace, name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/x-yaml", []byte(yaml))
}

// GetServiceYAML retourne le YAML d'un service.
func (h *K8sHandler) GetServiceYAML(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace et name requis"})
		return
	}

	yaml, err := h.svc.GetServiceYAML(ctx, namespace, name)
	if err != nil {
		log.Printf("Erreur GetServiceYAML pour service %s/%s: %v", namespace, name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/x-yaml", []byte(yaml))
}

// ScaleDeployment modifie le nombre de replicas d'un deployment.
func (h *K8sHandler) ScaleDeployment(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	name := c.Param("name")

	log.Printf("ScaleDeployment appelé pour %s/%s", namespace, name)

	// Accepter soit int32 soit float64 (JavaScript peut envoyer des nombres)
	var req struct {
		Replicas interface{} `json:"replicas" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Erreur de binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convertir en int32
	var replicas int32
	switch v := req.Replicas.(type) {
	case float64:
		replicas = int32(v)
	case int32:
		replicas = v
	case int:
		replicas = int32(v)
	case int64:
		replicas = int32(v)
	default:
		log.Printf("Type de replicas invalide: %T", req.Replicas)
		c.JSON(http.StatusBadRequest, gin.H{"error": "replicas doit être un nombre"})
		return
	}

	if replicas < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "replicas doit être >= 0"})
		return
	}

	log.Printf("ScaleDeployment: mise à jour de %s/%s à %d replicas", namespace, name, replicas)

	if err := h.svc.ScaleDeployment(ctx, namespace, name, replicas); err != nil {
		log.Printf("Erreur ScaleDeployment pour deployment %s/%s: %v", namespace, name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("ScaleDeployment réussi pour %s/%s", namespace, name)
	c.JSON(http.StatusOK, gin.H{"message": "deployment mis à jour", "replicas": replicas})
}

// DeletePod supprime un pod.
func (h *K8sHandler) DeletePod(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	name := c.Param("name")

	if err := h.svc.DeletePod(ctx, namespace, name); err != nil {
		log.Printf("Erreur DeletePod pour pod %s/%s: %v", namespace, name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pod supprimé"})
}

// DeleteDeployment supprime un deployment.
func (h *K8sHandler) DeleteDeployment(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	name := c.Param("name")

	if err := h.svc.DeleteDeployment(ctx, namespace, name); err != nil {
		log.Printf("Erreur DeleteDeployment pour deployment %s/%s: %v", namespace, name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deployment supprimé"})
}

// DeleteService supprime un service.
func (h *K8sHandler) DeleteService(c *gin.Context) {
	ctx := c.Request.Context()
	namespace := c.Param("namespace")
	name := c.Param("name")

	if err := h.svc.DeleteService(ctx, namespace, name); err != nil {
		log.Printf("Erreur DeleteService pour service %s/%s: %v", namespace, name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "service supprimé"})
}

// BulkDeleteRequest représente une requête pour supprimer plusieurs ressources.
type BulkDeleteRequest struct {
	Names []string `json:"names" binding:"required"`
}

// BulkDeleteResponse représente la réponse d'une action en masse.
type BulkDeleteResponse struct {
	Success []string          `json:"success"`
	Failed  map[string]string `json:"failed"`
	Total   int               `json:"total"`
}

// BulkDeletePods supprime plusieurs pods en une seule requête.
func (h *K8sHandler) BulkDeletePods(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	namespace := c.Param("namespace")
	ctx := c.Request.Context()

	var req BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format de requête invalide"})
		return
	}

	response := BulkDeleteResponse{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
		Total:   len(req.Names),
	}

	for _, name := range req.Names {
		if err := h.svc.DeletePod(ctx, namespace, name); err != nil {
			response.Failed[name] = err.Error()
			log.Printf("Erreur lors de la suppression du pod %s/%s: %v", namespace, name, err)
		} else {
			response.Success = append(response.Success, name)
		}
	}

	statusCode := http.StatusOK
	if len(response.Failed) > 0 {
		if len(response.Success) == 0 {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusMultiStatus // 207
		}
	}

	c.JSON(statusCode, response)
}

// BulkDeleteDeployments supprime plusieurs deployments en une seule requête.
func (h *K8sHandler) BulkDeleteDeployments(c *gin.Context) {
	namespace := c.Param("namespace")
	ctx := c.Request.Context()

	var req BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format de requête invalide"})
		return
	}

	response := BulkDeleteResponse{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
		Total:   len(req.Names),
	}

	for _, name := range req.Names {
		if err := h.svc.DeleteDeployment(ctx, namespace, name); err != nil {
			response.Failed[name] = err.Error()
			log.Printf("Erreur lors de la suppression du deployment %s/%s: %v", namespace, name, err)
		} else {
			response.Success = append(response.Success, name)
		}
	}

	statusCode := http.StatusOK
	if len(response.Failed) > 0 {
		if len(response.Success) == 0 {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusMultiStatus // 207
		}
	}

	c.JSON(statusCode, response)
}

// BulkDeleteServices supprime plusieurs services en une seule requête.
func (h *K8sHandler) BulkDeleteServices(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	namespace := c.Param("namespace")
	ctx := c.Request.Context()

	var req BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format de requête invalide"})
		return
	}

	response := BulkDeleteResponse{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
		Total:   len(req.Names),
	}

	for _, name := range req.Names {
		if err := h.svc.DeleteService(ctx, namespace, name); err != nil {
			response.Failed[name] = err.Error()
			log.Printf("Erreur lors de la suppression du service %s/%s: %v", namespace, name, err)
		} else {
			response.Success = append(response.Success, name)
		}
	}

	statusCode := http.StatusOK
	if len(response.Failed) > 0 {
		if len(response.Success) == 0 {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusMultiStatus // 207
		}
	}

	c.JSON(statusCode, response)
}

// BulkRestartPods redémarre plusieurs pods en supprimant et laissant Kubernetes les recréer.
func (h *K8sHandler) BulkRestartPods(c *gin.Context) {
	namespace := c.Param("namespace")
	ctx := c.Request.Context()

	var req BulkDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format de requête invalide"})
		return
	}

	response := BulkDeleteResponse{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
		Total:   len(req.Names),
	}

	for _, name := range req.Names {
		// Pour redémarrer un pod, on le supprime et Kubernetes le recrée automatiquement
		if err := h.svc.DeletePod(ctx, namespace, name); err != nil {
			response.Failed[name] = err.Error()
			log.Printf("Erreur lors du redémarrage du pod %s/%s: %v", namespace, name, err)
		} else {
			response.Success = append(response.Success, name)
		}
	}

	statusCode := http.StatusOK
	if len(response.Failed) > 0 {
		if len(response.Success) == 0 {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusMultiStatus // 207
		}
	}

	c.JSON(statusCode, response)
}

// BulkScaleRequest représente une requête pour scale plusieurs deployments.
type BulkScaleRequest struct {
	Names    []string `json:"names" binding:"required"`
	Replicas int32    `json:"replicas" binding:"required,min=0"`
}

// BulkScaleResponse représente la réponse d'une action de scale en masse.
type BulkScaleResponse struct {
	Success []string          `json:"success"`
	Failed  map[string]string `json:"failed"`
	Total   int               `json:"total"`
}

// BulkScaleDeployments scale plusieurs deployments en une seule requête.
func (h *K8sHandler) BulkScaleDeployments(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	namespace := c.Param("namespace")
	ctx := c.Request.Context()

	var req BulkScaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "format de requête invalide"})
		return
	}

	response := BulkScaleResponse{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
		Total:   len(req.Names),
	}

	for _, name := range req.Names {
		if err := h.svc.ScaleDeployment(ctx, namespace, name, req.Replicas); err != nil {
			response.Failed[name] = err.Error()
			log.Printf("Erreur lors du scale du deployment %s/%s: %v", namespace, name, err)
		} else {
			response.Success = append(response.Success, name)
		}
	}

	statusCode := http.StatusOK
	if len(response.Failed) > 0 {
		if len(response.Success) == 0 {
			statusCode = http.StatusInternalServerError
		} else {
			statusCode = http.StatusMultiStatus // 207
		}
	}

	c.JSON(statusCode, response)
}

// GetEvents retourne la liste des événements pour un namespace donné.
func (h *K8sHandler) GetEvents(c *gin.Context) {
	if !h.checkService(c) {
		return
	}
	ctx := c.Request.Context()
	namespace := c.Param("namespace")

	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace requis"})
		return
	}

	events, err := h.svc.ListEvents(ctx, namespace)
	if err != nil {
		log.Printf("Erreur GetEvents pour namespace %s: %v", namespace, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": events,
	})
}

