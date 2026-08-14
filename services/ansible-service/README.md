# ansible-service

Service Go d'intégration Ansible de Kura. Il s'appuie sur **Ansible Semaphore**
comme moteur d'exécution : Kura ne lance jamais `ansible-playbook` lui-même, il
pilote Semaphore par son API REST et traduit les réponses au format historique
AWX que consomme le frontend (aucun changement côté UI en cas de changement de
moteur).

## Architecture

```
internal/
  client/      Client REST Semaphore (traduction Semaphore → format AWX)
  config/      Chargement de la configuration (variables d'environnement)
  configstore/ Configuration persistée dans Postgres via l'auth-service
  handler/     Routes HTTP (Gin) + webhook
  parser/      Analyse YAML des playbooks (structure, statistiques)
  service/     Logique métier, cache Valkey, injection du cluster actif
  tracing/     OpenTelemetry → Tempo
```

## Configuration

| Variable | Défaut | Rôle |
|---|---|---|
| `ANSIBLE_SERVICE_PORT` | `8083` | Port d'écoute |
| `SEMAPHORE_URL` | — | URL de l'instance Semaphore |
| `SEMAPHORE_API_TOKEN` | — | Token API Semaphore |
| `SEMAPHORE_PROJECT_ID` | `1` | Projet Semaphore piloté |
| `VALKEY_HOST` / `VALKEY_PORT` / `VALKEY_PASSWORD` / `VALKEY_DB` | `localhost:6379` | Cache |
| `ANSIBLE_CACHE_TTL` | `300` | TTL du cache (secondes) |
| `AUTH_SERVICE_URL` | `http://auth-service:8080` | Configstore |
| `K8S_SERVICE_URL` | `http://k8s-service:8081` | Injection du cluster actif dans les jobs |
| `CODE_SERVICE_URL` | `http://code-service:8088` | Lecture du source des playbooks |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `tempo:4317` | Traces |

La configuration Semaphore saisie dans l'UI (Paramètres) est persistée en base
via le configstore et prime sur les variables d'environnement.

## API

Préfixe `/api/v1/ansible` : `jobs`, `jobs/history`, `jobs/:id`, `inventories`,
`job-templates` (+ `launch`, `playbook`), `projects`, `playbooks/analyze`,
`config`. Webhook : `POST /api/v1/webhooks/ansible`. Santé : `GET /health`.

## Développement

```bash
go build ./...
go test ./...
```
