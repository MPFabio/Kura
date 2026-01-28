"""Métriques Prometheus pour le service Ansible."""
import logging
from prometheus_client import Counter, Histogram, Gauge, generate_latest, CONTENT_TYPE_LATEST
from fastapi import Response

logger = logging.getLogger(__name__)

# Compteurs
api_requests_total = Counter(
    "ansible_api_requests_total",
    "Nombre total de requêtes API",
    ["method", "endpoint", "status"],
)

tower_api_requests_total = Counter(
    "ansible_tower_api_requests_total",
    "Nombre total de requêtes vers Ansible Tower",
    ["method", "endpoint", "status"],
)

jobs_launched_total = Counter(
    "ansible_jobs_launched_total",
    "Nombre total de jobs lancés",
    ["template_id"],
)

webhooks_received_total = Counter(
    "ansible_webhooks_received_total",
    "Nombre total de webhooks reçus",
    ["event_type"],
)

# Histogrammes
api_request_duration_seconds = Histogram(
    "ansible_api_request_duration_seconds",
    "Durée des requêtes API en secondes",
    ["method", "endpoint"],
    buckets=[0.1, 0.5, 1.0, 2.0, 5.0, 10.0],
)

tower_api_request_duration_seconds = Histogram(
    "ansible_tower_api_request_duration_seconds",
    "Durée des requêtes vers Ansible Tower en secondes",
    ["method", "endpoint"],
    buckets=[0.1, 0.5, 1.0, 2.0, 5.0, 10.0],
)

# Jauges
active_websocket_connections = Gauge(
    "ansible_active_websocket_connections",
    "Nombre de connexions WebSocket actives",
)

cache_hits_total = Counter(
    "ansible_cache_hits_total",
    "Nombre total de hits de cache",
    ["cache_type"],
)

cache_misses_total = Counter(
    "ansible_cache_misses_total",
    "Nombre total de misses de cache",
    ["cache_type"],
)


def get_metrics_response() -> Response:
    """Retourne la réponse avec les métriques Prometheus."""
    return Response(
        content=generate_latest(),
        media_type=CONTENT_TYPE_LATEST,
    )
