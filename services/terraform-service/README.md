# Service Terraform

Service de gestion des états Terraform avec détection de drift et API REST.

## Fonctionnalités

- **Parser tfstate** : Parse les fichiers tfstate (format JSON)
- **Détection de drift** : Détecte les différences entre l'état Terraform et l'état réel
- **API REST** : API complète pour gérer les états Terraform
- **Cache Redis** : Mise en cache pour améliorer les performances

## Prérequis

- Go 1.21+
- Redis 7+ (pour le cache)
- Docker (optionnel, pour la containerisation)

## Installation

### Local

```bash
# Installer les dépendances
go mod download

# Lancer le service
go run main.go
```

### Docker

```bash
# Builder l'image
docker build -t terraform-service .

# Lancer le conteneur
docker run -p 8082:8082 \
  -e REDIS_HOST=localhost \
  -e REDIS_PORT=6379 \
  terraform-service
```

## Configuration

Le service utilise les variables d'environnement suivantes :

- `TERRAFORM_SERVICE_PORT` : Port du serveur (défaut: 8082)
- `REDIS_HOST` : Host Redis (défaut: localhost)
- `REDIS_PORT` : Port Redis (défaut: 6379)
- `REDIS_PASSWORD` : Mot de passe Redis (optionnel)
- `REDIS_DB` : Base de données Redis (défaut: 0)
- `TERRAFORM_CACHE_TTL` : TTL du cache (défaut: 5m)
- `ENV` : Environnement (development/production)
- `LOG_LEVEL` : Niveau de log (info/debug)

## API

### Santé

- `GET /health` - Vérifier l'état du service

### États Terraform

- `POST /api/v1/terraform/states/upload` - Uploader un fichier tfstate (multipart/form-data)
- `POST /api/v1/terraform/states` - Uploader un tfstate (JSON)
- `GET /api/v1/terraform/states` - Lister tous les états
- `GET /api/v1/terraform/states/:id` - Récupérer un état par ID
- `GET /api/v1/terraform/states/:id/summary` - Résumé d'un état
- `DELETE /api/v1/terraform/states/:id` - Supprimer un état

### Ressources

- `GET /api/v1/terraform/states/:id/resources` - Lister toutes les ressources d'un état
- `GET /api/v1/terraform/states/:id/resources/:address` - Récupérer une ressource par adresse

### Sorties

- `GET /api/v1/terraform/states/:id/outputs` - Lister toutes les sorties d'un état

### Détection de drift

- `POST /api/v1/terraform/states/:id/drift` - Détecter les dérives pour un état

## Exemples

### Uploader un tfstate

```bash
# Via fichier
curl -X POST http://localhost:8082/api/v1/terraform/states/upload \
  -F "file=@terraform.tfstate" \
  -F "name=my-terraform-state"

# Via JSON
curl -X POST http://localhost:8082/api/v1/terraform/states \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-terraform-state",
    "state": {...}
  }'
```

### Lister les états

```bash
curl http://localhost:8082/api/v1/terraform/states
```

### Récupérer un état

```bash
curl http://localhost:8082/api/v1/terraform/states/{id}
```

### Détecter le drift

```bash
curl -X POST http://localhost:8082/api/v1/terraform/states/{id}/drift
```

## Structure du code

```
terraform-service/
├── main.go                          # Point d'entrée
├── Dockerfile                       # Image Docker
├── go.mod                          # Dépendances
├── internal/
│   ├── config/                     # Configuration
│   ├── cache/                      # Cache Redis
│   ├── handler/                    # Handlers HTTP
│   ├── models/                     # Modèles de données
│   ├── parser/                     # Parser tfstate
│   └── service/                    # Logique métier
└── README.md                       # Documentation
```

## Développement futur

- [ ] Intégration avec Terraform Cloud/Enterprise API
- [ ] Détection de drift avancée avec providers réels (AWS, GCP, Azure)
- [ ] Queue Kafka pour traitement asynchrone
- [ ] Persistance dans PostgreSQL
- [ ] Webhooks pour notifications
- [ ] Historique des changements
