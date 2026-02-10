#!/bin/bash
# Script de test du pipeline-service
# Prérequis: Redis actif, service démarré sur le port 8084

BASE_URL="${PIPELINE_SERVICE_URL:-http://localhost:8084}"

echo "=== Test Pipeline Service ==="
echo "URL: $BASE_URL"
echo ""

echo "1. Health check"
curl -s "$BASE_URL/health" | jq . 2>/dev/null || curl -s "$BASE_URL/health"
echo -e "\n"

echo "2. Liste des providers"
curl -s "$BASE_URL/api/v1/pipeline/providers" | jq . 2>/dev/null || curl -s "$BASE_URL/api/v1/pipeline/providers"
echo -e "\n"

echo "3. Liste des runs (vide)"
curl -s "$BASE_URL/api/v1/pipeline/runs?limit=5" | jq . 2>/dev/null || curl -s "$BASE_URL/api/v1/pipeline/runs?limit=5"
echo -e "\n"

echo "4. Webhook GitHub (payload minimal)"
curl -s -X POST "$BASE_URL/api/v1/pipeline/webhooks/github" \
  -H "Content-Type: application/json" \
  -d '{
    "action": "completed",
    "workflow_run": {
      "id": 123,
      "name": "CI",
      "status": "completed",
      "conclusion": "success",
      "html_url": "https://github.com/org/repo/actions/runs/123",
      "run_started_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:05:00Z",
      "head_branch": "main",
      "head_sha": "abc123",
      "display_title": "Test",
      "repository": {"full_name": "org/repo"},
      "triggering_actor": {"login": "user"}
    }
  }' | jq . 2>/dev/null || echo " (sans jq)"
echo -e "\n"

echo "5. Liste des runs (après webhook)"
curl -s "$BASE_URL/api/v1/pipeline/runs?provider=github&repository=org/repo&branch=main&limit=5" | jq . 2>/dev/null || curl -s "$BASE_URL/api/v1/pipeline/runs?provider=github&repository=org/repo&branch=main&limit=5"
echo -e "\n"

echo "6. Statut agrégé"
curl -s "$BASE_URL/api/v1/pipeline/aggregated?provider=github&repository=org/repo&branch=main" | jq . 2>/dev/null || curl -s "$BASE_URL/api/v1/pipeline/aggregated?provider=github&repository=org/repo&branch=main"
echo -e "\n"

echo "=== Tests terminés ==="
