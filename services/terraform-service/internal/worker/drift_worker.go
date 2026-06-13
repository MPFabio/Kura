package worker

import (
	"context"
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
	cancel           context.CancelFunc
	done             chan struct{}
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
		done:             make(chan struct{}),
		interval:         interval,
	}
}

// Start démarre le worker dans une goroutine. Le contexte racine est stocké
// pour être annulé lors de l'arrêt, ce qui propage l'annulation à tous les
// appels DetectDrift en cours.
func (w *DriftWorker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	log.Printf("🔄 Drift worker démarré (intervalle: %s)", w.interval)
	go w.run(ctx)
}

// Stop demande l'arrêt gracieux du worker et attend sa terminaison.
func (w *DriftWorker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	<-w.done
	log.Println("✅ Drift worker arrêté proprement")
}

func (w *DriftWorker) run(ctx context.Context) {
	defer close(w.done)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Premier run après un court délai (laisse le temps aux services de démarrer).
	select {
	case <-time.After(30 * time.Second):
		w.runDriftCheck(ctx)
	case <-ctx.Done():
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.runDriftCheck(ctx)
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
		// Vérifier l'annulation entre chaque source pour ne pas bloquer l'arrêt.
		select {
		case <-ctx.Done():
			return
		default:
		}

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

func (w *DriftWorker) emitDriftEvent(ctx context.Context, stateFileID, sourceID string, results []*models.DriftResult) {
	driftCount := 0
	for _, r := range results {
		if r.Status != "in_sync" && r.Status != "unknown" {
			driftCount++
		}
	}

	log.Printf("🔔 Drift détecté: state=%s source=%s count=%d", stateFileID, sourceID, driftCount)
}
