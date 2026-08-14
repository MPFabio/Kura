package handler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Métriques métier exposées sur /metrics (scrapées par vmagent), en complément
// des métriques go_* et process_* du registre par défaut.
var (
	jobsLaunchedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ansible_jobs_launched_total",
		Help: "Nombre de jobs Ansible lancés depuis un template.",
	}, []string{"template_id"})

	webhooksReceivedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ansible_webhooks_received_total",
		Help: "Nombre de webhooks Ansible reçus, par type d'événement.",
	}, []string{"event"})
)
