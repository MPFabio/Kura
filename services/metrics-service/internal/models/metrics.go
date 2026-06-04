package models

// ServiceHealth représente l'état de santé d'un microservice Kura.
type ServiceHealth struct {
	Name       string  `json:"name"`
	Job        string  `json:"job"`
	Up         bool    `json:"up"`
	Goroutines float64 `json:"goroutines"`
	MemoryMB   float64 `json:"memory_mb"`
}

// ServiceMetric représente les métriques détaillées d'un service.
type ServiceMetric struct {
	Name        string  `json:"name"`
	Job         string  `json:"job"`
	Up          bool    `json:"up"`
	Goroutines  float64 `json:"goroutines"`
	CPURate     float64 `json:"cpu_rate"`
	MemoryMB    float64 `json:"memory_mb"`
}

// Overview représente les KPIs globaux de la plateforme.
type Overview struct {
	TotalServices   int     `json:"total_services"`
	ServicesUp      int     `json:"services_up"`
	ServicesDown    int     `json:"services_down"`
	TotalGoroutines float64 `json:"total_goroutines"`
	TotalMemoryMB   float64 `json:"total_memory_mb"`
}

// PrometheusResponse est la réponse brute de l'API Prometheus.
type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}
