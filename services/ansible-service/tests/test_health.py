"""Tests basiques pour l'endpoint /health et /metrics du service Ansible."""

from fastapi.testclient import TestClient

from main import app


client = TestClient(app)


def test_health_endpoint_returns_ok_and_flags():
    response = client.get("/health")
    assert response.status_code == 200
    body = response.json()
    assert body.get("status") == "ok"
    assert body.get("service") == "ansible-service"
    # Clés de diagnostic présentes
    assert "ansible_tower_configured" in body
    assert "redis_available" in body


def test_metrics_endpoint_exposes_prometheus_metrics():
    response = client.get("/metrics")
    assert response.status_code == 200
    assert "text/plain" in response.headers["content-type"]
    # Le contenu doit contenir au moins une métrique définie dans internal.metrics.prometheus
    assert b"ansible_api_requests_total" in response.content
