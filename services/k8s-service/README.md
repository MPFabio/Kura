# K8s Service - Service Kubernetes ModulOps

Service microservices en Go exposant des informations Kubernetes via une API REST, avec cache Redis et support de webhooks.

## Fonctionnalités

- Listing des namespaces Kubernetes
- Listing des pods par namespace
- Cache Redis des réponses pour réduire la charge sur l'API K8s
- Endpoint webhook générique pour recevoir des événements liés à Kubernetes

## Structure

```
k8s-service/
├── main.go                    # Point d'entrée
├── go.mod                     # Dépendances Go
├── Dockerfile                 # Image Docker
├── internal/
│   ├── config/                # Configuration (env, Redis, Kubeconfig)
│   ├── k8s/                   # Client Kubernetes
│   ├── cache/                 # Client Redis
│   ├── service/               # Logique métier (K8s + cache)
│   └── handler/               # Handlers HTTP / API REST
└── README.md
```

## Configuration

Variables d'environnement principales :

- `K8S_SERVICE_PORT` : Port HTTP du service (défaut: 8081)
- `ENV` : Environnement (`development`, `production`)
- `LOG_LEVEL` : Niveau de log (info, debug, error)

### Kubernetes

- `K8S_INCLUSTER` : `true` pour utiliser la configuration in-cluster, `false` pour utiliser un kubeconfig (défaut: false)
- `KUBECONFIG_PATH` : Chemin vers le fichier kubeconfig quand `K8S_INCLUSTER=false`

### Redis

- `REDIS_HOST` : Hôte Redis (défaut: localhost)
- `REDIS_PORT` : Port Redis (défaut: 6379)
- `REDIS_PASSWORD` : Mot de passe Redis (optionnel)
- `REDIS_DB` : Index de base de données Redis (défaut: 0)
- `K8S_CACHE_TTL` : TTL des entrées de cache (défaut: 30s)

## API Endpoints

### Santé

- `GET /health`  
  Retourne l'état basique du service.

### Kubernetes

- `GET /api/v1/k8s/namespaces`  
  Liste les namespaces Kubernetes (avec cache Redis).

- `GET /api/v1/k8s/namespaces/:namespace/pods`  
  Liste les pods d'un namespace donné (avec cache Redis).

### Webhooks

- `POST /api/v1/k8s/webhooks/events`  
  Endpoint générique pour recevoir des événements Kubernetes (payload JSON arbitraire, actuellement simplement loggué).

## Développement local

```bash
cd services/k8s-service

# Télécharger les dépendances
go mod download

# Exporter les variables nécessaires, par exemple :
export K8S_INCLUSTER=false
export KUBECONFIG_PATH=$HOME/.kube/config
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Lancer le service
go run main.go
```

## Build Docker

```bash
docker build -t modulops-k8s-service .
```

