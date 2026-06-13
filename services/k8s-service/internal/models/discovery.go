package models

// DiscoveredComponent représente un composant de la stack d'observabilité du
// cluster client détecté automatiquement par recherche de labels (Prometheus
// ou VictoriaMetrics, Grafana, Loki, Tempo).
type DiscoveredComponent struct {
	Name      string `json:"name"`
	Found     bool   `json:"found"`
	Namespace string `json:"namespace,omitempty"`
	PodName   string `json:"pod_name,omitempty"`
}

// DiscoveryReport résume ce que Kura a détecté automatiquement dans le
// cluster client : applications ArgoCD déployées et composants
// d'observabilité reconnus parmi elles.
type DiscoveryReport struct {
	Applications []ArgoApplication      `json:"applications"`
	Observability []DiscoveredComponent `json:"observability"`
}
