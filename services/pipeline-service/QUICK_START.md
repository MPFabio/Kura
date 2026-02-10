# Pipeline Service - Démarrage rapide

## Prérequis

- Go 1.21+
- Redis (ou Docker)

## Démarrage

```bash
# Depuis la racine du projet
docker compose up -d redis
docker compose up -d pipeline-service

# Ou en local
export REDIS_HOST=localhost
export REDIS_PORT=6379
go run main.go
```

## Tester le service

```bash
# Santé
curl http://localhost:8084/health

# Liste des providers
curl http://localhost:8084/api/v1/pipeline/providers

# Liste des runs (vide au départ)
curl "http://localhost:8084/api/v1/pipeline/runs?limit=10"

# Webhook GitHub (exemple de payload minimal)
curl -X POST http://localhost:8084/api/v1/pipeline/webhooks/github \
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
      "display_title": "Fix bug",
      "repository": {"full_name": "org/repo"},
      "triggering_actor": {"login": "user"}
    }
  }'
```

## Configuration des webhooks

Configurez les webhooks dans vos projets GitHub/GitLab/Jenkins pour pointer vers :
- `http://localhost:8000/api/v1/pipeline/webhooks/github` (via Kong)
- `http://localhost:8000/api/v1/pipeline/webhooks/gitlab`
- `http://localhost:8000/api/v1/pipeline/webhooks/jenkins`
