# Pipeline Service

Service d'agrégation et de gestion des pipelines CI/CD (GitHub Actions, GitLab CI, Jenkins).

## Fonctionnalités

- **Adapters** : GitHub Actions, GitLab CI, Jenkins
- **Webhooks** : Réception des événements en temps réel depuis chaque provider
- **Agrégation des statuts** : Vue consolidée par repository/branch
- **Historique des exécutions** : Stockage et consultation des runs
- **Métriques Prometheus** : Compteurs et histogrammes

## API

### Endpoints REST

| Méthode | Endpoint | Description |
|---------|----------|-------------|
| GET | `/api/v1/pipeline/runs` | Liste les exécutions (filtres: provider, repository, branch, limit) |
| GET | `/api/v1/pipeline/runs/:id` | Détail d'une exécution |
| GET | `/api/v1/pipeline/aggregated` | Statut agrégé (provider, repository, branch requis) |
| GET | `/api/v1/pipeline/providers` | Liste des providers supportés |

### Webhooks

| Provider | Endpoint | Headers |
|----------|----------|---------|
| GitHub | POST `/api/v1/pipeline/webhooks/github` | X-GitHub-Event (optionnel) |
| GitLab | POST `/api/v1/pipeline/webhooks/gitlab` | X-Gitlab-Event (optionnel) |
| Jenkins | POST `/api/v1/pipeline/webhooks/jenkins` | - |
| Auto | POST `/api/v1/pipeline/webhooks` | Détection via X-GitHub-Event ou X-Gitlab-Event |

## Configuration

Variables d'environnement :

| Variable | Défaut | Description |
|----------|--------|-------------|
| PIPELINE_SERVICE_PORT | 8084 | Port du serveur |
| REDIS_HOST | localhost | Hôte Redis |
| REDIS_PORT | 6379 | Port Redis |
| PIPELINE_CACHE_TTL | 24h | TTL du cache des runs |
| GITHUB_TOKEN | - | Token API GitHub (optionnel) |
| GITLAB_TOKEN | - | Token API GitLab (optionnel) |
| JENKINS_URL | - | URL Jenkins (optionnel) |

## Démarrage local

```bash
# Avec Docker Compose (depuis la racine du projet)
docker compose up -d pipeline-service

# Ou en standalone (Redis requis)
cd services/pipeline-service
go run main.go
```

## Webhooks - Configuration chez les providers

### GitHub Actions
1. Settings > Webhooks > Add webhook
2. Payload URL : `https://votre-domain/api/v1/pipeline/webhooks/github`
3. Content type : application/json
4. Events : Workflow runs (ou Check runs)

### GitLab CI
1. Settings > Webhooks
2. URL : `https://votre-domain/api/v1/pipeline/webhooks/gitlab`
3. Trigger : Pipeline events, Job events

### Jenkins
1. Post-build Actions > Notify external system
2. Ou utiliser le plugin "Generic Webhook Trigger"
3. URL : `https://votre-domain/api/v1/pipeline/webhooks/jenkins`
