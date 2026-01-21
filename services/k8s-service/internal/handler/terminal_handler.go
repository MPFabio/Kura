package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/modulops/k8s-service/internal/config"
	"github.com/modulops/k8s-service/internal/k8s"
	"github.com/modulops/k8s-service/internal/service"
	"k8s.io/client-go/tools/remotecommand"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Permettre toutes les origines en développement
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// TerminalHandler gère les connexions WebSocket pour le terminal.
type TerminalHandler struct {
	svc            *service.K8sService
	clusterService *service.ClusterService
	redisClient    service.Cache
	cfg            *config.Config
	mu             sync.RWMutex
}

// NewTerminalHandler crée un nouveau handler de terminal.
func NewTerminalHandler(svc *service.K8sService, clusterService *service.ClusterService, redisClient service.Cache, cfg *config.Config) *TerminalHandler {
	return &TerminalHandler{
		svc:            svc,
		clusterService: clusterService,
		redisClient:    redisClient,
		cfg:            cfg,
	}
}

// TerminalMessage représente un message du terminal.
type TerminalMessage struct {
	Type      string `json:"type"`      // "resize", "input", "init"
	Data      string `json:"data"`      // Données (commande, input, etc.)
	Namespace string `json:"namespace"` // Namespace du pod
	Pod       string `json:"pod"`       // Nom du pod
	Container string `json:"container"` // Container (optionnel)
	Width     int    `json:"width"`     // Largeur du terminal
	Height    int    `json:"height"`    // Hauteur du terminal
}

// getService obtient le service K8s, en le créant dynamiquement si nécessaire.
func (h *TerminalHandler) getService(ctx context.Context) (*service.K8sService, error) {
	h.mu.RLock()
	svc := h.svc
	h.mu.RUnlock()

	// Si le service existe déjà, on l'utilise
	if svc != nil {
		return svc, nil
	}

	// Sinon, essayer de créer le service à partir du cluster actif
	if h.clusterService == nil {
		return nil, fmt.Errorf("aucun cluster Kubernetes configuré")
	}

	activeCluster, err := h.clusterService.GetActiveCluster(ctx)
	if err != nil {
		return nil, fmt.Errorf("aucun cluster Kubernetes actif: %w", err)
	}

	// Créer un client Kubernetes temporaire à partir du cluster actif
	kubeconfigPath, err := h.clusterService.SaveKubeconfigToFile(ctx, activeCluster.ID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la préparation du kubeconfig: %w", err)
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
		return nil, fmt.Errorf("erreur lors de la connexion au cluster: %w", err)
	}

	// Créer le service avec le nouveau client
	var k8sClientInterface service.K8sClient = k8sClient
	newSvc := service.NewK8sService(k8sClientInterface, h.redisClient, h.cfg)

	// Mettre à jour le service de manière thread-safe
	h.mu.Lock()
	h.svc = newSvc
	h.mu.Unlock()

	return newSvc, nil
}

// HandleTerminal gère la connexion WebSocket pour le terminal.
func (h *TerminalHandler) HandleTerminal(c *gin.Context) {
	// Récupérer les paramètres depuis l'URL
	namespace := c.Param("namespace")
	pod := c.Param("name")
	
	if namespace == "" || pod == "" {
		log.Printf("Paramètres manquants: namespace=%s, pod=%s", namespace, pod)
		c.JSON(400, gin.H{"error": "namespace et name requis"})
		return
	}

	log.Printf("Connexion terminal demandée pour pod %s/%s", namespace, pod)

	// Obtenir le service (créé dynamiquement si nécessaire)
	ctx := c.Request.Context()
	svc, err := h.getService(ctx)
	if err != nil {
		log.Printf("Erreur lors de l'obtention du service: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": err.Error(),
		})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Erreur lors de l'upgrade WebSocket pour pod %s/%s: %v", namespace, pod, err)
		// Ne pas utiliser c.JSON() après un échec d'upgrade, la réponse peut déjà être partiellement écrite
		return
	}
	defer conn.Close()

	log.Printf("WebSocket connecté pour pod %s/%s", namespace, pod)

	// Lire le message initial pour obtenir les paramètres (container, taille, etc.)
	var initMsg TerminalMessage
	if err := conn.ReadJSON(&initMsg); err != nil {
		log.Printf("Erreur lors de la lecture du message initial: %v", err)
		return
	}

	if initMsg.Type != "init" {
		log.Printf("Message initial invalide: %v", initMsg)
		return
	}

	container := initMsg.Container
	if container == "" {
		container = "" // Utiliser le container par défaut
	}

	// Créer les pipes pour stdin/stdout/stderr
	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	// Créer un sizeQueue partagé pour le resize du terminal
	sizeQueue := &terminalSizeQueue{
		sizeChan: make(chan remotecommand.TerminalSize, 1),
	}
	// Initialiser la taille du terminal
	sizeQueue.Resize(80, 24)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(3)

	// Goroutine pour lire stdout et l'envoyer au WebSocket
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			n, err := stdoutReader.Read(buf)
			if n > 0 {
				if err := conn.WriteJSON(map[string]interface{}{
					"type": "output",
					"data": string(buf[:n]),
				}); err != nil {
					log.Printf("Erreur lors de l'écriture stdout: %v", err)
					cancel()
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("Erreur lors de la lecture stdout: %v", err)
				}
				return
			}
		}
	}()

	// Goroutine pour lire stderr et l'envoyer au WebSocket
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			n, err := stderrReader.Read(buf)
			if n > 0 {
				if err := conn.WriteJSON(map[string]interface{}{
					"type": "error",
					"data": string(buf[:n]),
				}); err != nil {
					log.Printf("Erreur lors de l'écriture stderr: %v", err)
					cancel()
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("Erreur lors de la lecture stderr: %v", err)
				}
				return
			}
		}
	}()

	// Goroutine pour exécuter la commande dans le pod
	go func() {
		defer wg.Done()
		defer stdinWriter.Close()
		defer stdoutWriter.Close()
		defer stderrWriter.Close()

		if container == "" {
			// Essayer de trouver le premier container
			podInfo, err := svc.GetPod(ctx, namespace, pod)
			if err == nil && len(podInfo.Spec.Containers) > 0 {
				container = podInfo.Spec.Containers[0].Name
			}
		}

		// Essayer /bin/sh d'abord (le plus commun)
		command := []string{"/bin/sh"}
		log.Printf("Exécution de /bin/sh dans le pod %s/%s", namespace, pod)
		
		err := svc.ExecPod(ctx, namespace, pod, container, command, stdinReader, stdoutWriter, stderrWriter, true, sizeQueue)
		if err != nil {
			log.Printf("Erreur lors de l'exécution de /bin/sh dans le pod %s/%s: %v", namespace, pod, err)
			
			errStr := err.Error()
			var errorMsg string
			
			// Si l'erreur indique que /bin/sh n'existe pas, suggérer d'essayer d'autres shells ou conteneurs
			if contains(errStr, "no such file") || contains(errStr, "not found") || contains(errStr, "stat") {
				errorMsg = fmt.Sprintf("Le shell /bin/sh n'est pas disponible dans ce conteneur.\r\n") +
					fmt.Sprintf("Ce conteneur semble être une image minimale (comme coredns) sans shell interactif.\r\n") +
					fmt.Sprintf("Suggestions:\r\n") +
					fmt.Sprintf("  - Essayez un autre pod qui contient un shell (par exemple, un pod avec une image basée sur Ubuntu, Alpine, ou Debian)\r\n") +
					fmt.Sprintf("  - Certains conteneurs utilisent /bin/bash ou /bin/ash au lieu de /bin/sh\r\n") +
					fmt.Sprintf("  - Les images minimales comme coredns, pause, ou distroless n'ont généralement pas de shell\r\n")
			} else {
				errorMsg = "Erreur lors de l'exécution: " + err.Error()
			}
			
			// Envoyer l'erreur au client
			if writeErr := conn.WriteJSON(map[string]interface{}{
				"type": "error",
				"data": errorMsg + "\r\n",
			}); writeErr != nil {
				log.Printf("Erreur lors de l'écriture de l'erreur: %v", writeErr)
			}
		}
	}()

	// Lire les messages du WebSocket et les envoyer à stdin
	for {
		var msg TerminalMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Erreur WebSocket: %v", err)
			}
			break
		}

		switch msg.Type {
		case "input":
			if _, err := stdinWriter.Write([]byte(msg.Data)); err != nil {
				log.Printf("Erreur lors de l'écriture stdin: %v", err)
				return
			}
		case "resize":
			sizeQueue.Resize(uint16(msg.Width), uint16(msg.Height))
		}
	}

	cancel()
	wg.Wait()
}

// terminalSizeQueue implémente TerminalSizeQueue pour le resize du terminal.
type terminalSizeQueue struct {
	sizeChan chan remotecommand.TerminalSize
}

func (t *terminalSizeQueue) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.sizeChan:
		return &size
	default:
		return nil
	}
}

func (t *terminalSizeQueue) Resize(width, height uint16) {
	select {
	case t.sizeChan <- remotecommand.TerminalSize{Width: width, Height: height}:
	default:
	}
}

// contains vérifie si une chaîne contient une sous-chaîne (insensible à la casse)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
