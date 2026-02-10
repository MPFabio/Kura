package worker

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/modulops/terraform-service/internal/config"
	"github.com/modulops/terraform-service/internal/models"
	"github.com/modulops/terraform-service/internal/service"
)

const (
	defaultDriftInterval = 1 * time.Hour
)

// DriftWorker exécute la détection de drift en arrière-plan pour les sources avec auto_sync.
type DriftWorker struct {
	terraformService *service.TerraformService
	syncService      *service.SyncService
	cfg              *config.Config
	stopChan         chan struct{}
	interval         time.Duration
}

// NewDriftWorker crée un nouveau worker de drift.
func NewDriftWorker(terraformService *service.TerraformService, syncService *service.SyncService, cfg *config.Config) *DriftWorker {
	interval := defaultDriftInterval
	if cfg != nil && cfg.DriftWorkerInterval > 0 {
		interval = cfg.DriftWorkerInterval
	}
	return &DriftWorker{
		terraformService: terraformService,
		syncService:      syncService,
		cfg:              cfg,
		stopChan:         make(chan struct{}),
		interval:         interval,
	}
}

// Start démarre le worker.
func (w *DriftWorker) Start() {
	log.Printf("🔄 Drift worker démarré (intervalle: %s)", w.interval)
	go w.run()
}

// Stop arrête le worker.
func (w *DriftWorker) Stop() {
	close(w.stopChan)
	log.Println("Drift worker arrêté")
}

func (w *DriftWorker) run() {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Premier run après un court délai
	select {
	case <-time.After(30 * time.Second):
		w.runDriftCheck(context.Background())
	case <-w.stopChan:
		return
	}

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.runDriftCheck(context.Background())
		}
	}
}

func (w *DriftWorker) runDriftCheck(ctx context.Context) {
	sources, err := w.syncService.ListSources(ctx)
	if err != nil {
		log.Printf("⚠️  Drift worker: erreur ListSources: %v", err)
		return
	}

	for _, source := range sources {
		if !source.Enabled || !source.Config.AutoSync {
			continue
		}

		var credentialsJSON string
		providerType := source.Type
		if source.Type == "gcp" && source.Config.GCPCredentialsJSON != "" {
			sourceCopy := *source
			if err := w.syncService.DecryptCredentials(&sourceCopy); err == nil {
				credentialsJSON = sourceCopy.Config.GCPCredentialsJSON
			}
		}

		results, err := w.terraformService.DetectDrift(ctx, source.StateFileID, credentialsJSON, providerType)
		if err != nil {
			log.Printf("⚠️  Drift worker: état %s: %v", source.StateFileID, err)
			continue
		}

		drifted := false
		for _, r := range results {
			if r.Status != "in_sync" && r.Status != "unknown" {
				drifted = true
				break
			}
		}

		if drifted {
			w.emitDriftEvent(ctx, source.StateFileID, source.ID, results)
		}
	}
}

// DriftEventPayload représente le payload Kafka terraform.drift.detected.
type DriftEventPayload struct {
	EventType    string               `json:"event_type"`
	StateFileID  string               `json:"state_file_id"`
	SourceID     string               `json:"source_id"`
	DetectedAt   time.Time            `json:"detected_at"`
	DriftCount   int                  `json:"drift_count"`
	Results      []*models.DriftResult `json:"results,omitempty"`
}

func (w *DriftWorker) emitDriftEvent(ctx context.Context, stateFileID, sourceID string, results []*models.DriftResult) {
	driftCount := 0
	for _, r := range results {
		if r.Status != "in_sync" && r.Status != "unknown" {
			driftCount++
		}
	}

	payload := DriftEventPayload{
		EventType:   "terraform.drift.detected",
		StateFileID: stateFileID,
		SourceID:    sourceID,
		DetectedAt:  time.Now(),
		DriftCount:  driftCount,
		Results:     results,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("⚠️  Drift worker: marshal event: %v", err)
		return
	}

	// Publier sur Kafka si configuré (pour alimenter les alertes)
	if w.cfg != nil && w.cfg.KafkaBrokers != "" {
		if err := w.publishKafka(ctx, body); err != nil {
			log.Printf("⚠️  Drift worker: publish Kafka: %v", err)
		}
	}

	log.Printf("🔔 Drift détecté: state=%s source=%s count=%d", stateFileID, sourceID, driftCount)
}

// publishKafka publie l'événement sur Kafka (implémentation basique via interface si disponible).
func (w *DriftWorker) publishKafka(ctx context.Context, body []byte) error {
	// Kafka producer optionnel - pour l'instant on log seulement
	// Une implémentation avec segmentio/kafka-go peut être ajoutée si KAFKA_BROKERS est défini
	_ = ctx
	_ = body
	return nil
}
