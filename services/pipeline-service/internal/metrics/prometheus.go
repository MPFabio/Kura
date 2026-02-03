package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// WebhooksReceivedTotal compte les webhooks reçus par provider
	WebhooksReceivedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pipeline_webhooks_received_total",
			Help: "Nombre total de webhooks reçus",
		},
		[]string{"provider", "status"},
	)

	// WebhookProcessingDuration mesure la durée de traitement des webhooks
	WebhookProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pipeline_webhook_processing_duration_seconds",
			Help:    "Durée de traitement des webhooks en secondes",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12),
		},
		[]string{"provider"},
	)

	// RunsStoredTotal compte les runs stockés
	RunsStoredTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pipeline_runs_stored_total",
			Help: "Nombre total de runs stockés",
		},
		[]string{"provider", "status"},
	)

	// APIRequestsTotal compte les requêtes API
	APIRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pipeline_api_requests_total",
			Help: "Nombre total de requêtes API",
		},
		[]string{"method", "endpoint", "status"},
	)

	// APIRequestDuration mesure la durée des requêtes API
	APIRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "pipeline_api_request_duration_seconds",
			Help:    "Durée des requêtes API en secondes",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "endpoint"},
	)
)
