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

// LogEntry représente une ligne de log issue de Loki.
type LogEntry struct {
	Timestamp string            `json:"timestamp"`
	Line      string            `json:"line"`
	Labels    map[string]string `json:"labels"`
}

// LokiResponse est la réponse brute de l'API de requête Loki.
type LokiResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Stream map[string]string `json:"stream"`
			Values [][2]string        `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// TraceSummary représente un résumé de trace retourné par la recherche Tempo.
type TraceSummary struct {
	TraceID           string `json:"trace_id"`
	RootServiceName   string `json:"root_service_name"`
	RootTraceName     string `json:"root_trace_name"`
	StartTimeUnixNano string `json:"start_time_unix_nano"`
	DurationMs        int64  `json:"duration_ms"`
}

// TempoSearchResponse est la réponse brute de l'API de recherche Tempo (/api/search).
type TempoSearchResponse struct {
	Traces []struct {
		TraceID           string `json:"traceID"`
		RootServiceName   string `json:"rootServiceName"`
		RootTraceName     string `json:"rootTraceName"`
		StartTimeUnixNano string `json:"startTimeUnixNano"`
		DurationMs        int64  `json:"durationMs"`
	} `json:"traces"`
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
