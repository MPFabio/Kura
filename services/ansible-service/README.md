# Service Ansible Tower

Service REST Python pour interagir avec l'API Ansible Tower (AWX).

## Fonctionnalités

- **Client REST Ansible Tower** : Client complet pour l'API REST d'Ansible Tower
- **Gestion des jobs** : Récupération des jobs, détails, historique et sortie standard
- **Gestion des inventaires** : Liste des inventaires, détails et hôtes
- **Templates de jobs** : Récupération et lancement de templates de jobs
- **Projets et playbooks** : Gestion des projets et analyse des playbooks
- **Cache Redis** : Mise en cache pour améliorer les performances
- **API REST** : API complète avec FastAPI et documentation automatique

## Prérequis

- Python 3.11+
- Redis 7+ (pour le cache)
- Ansible Tower ou AWX (optionnel pour le développement)
- Docker (optionnel, pour la containerisation)

## Installation

### Local

```bash
# Installer les dépendances
pip install -r requirements.txt

# Lancer le service
python main.py

# Ou avec uvicorn directement
uvicorn main:app --host 0.0.0.0 --port 8083
```

### Docker

```bash
# Builder l'image
docker build -t ansible-service .

# Lancer le conteneur
docker run -p 8083:8083 \
  -e REDIS_HOST=localhost \
  -e REDIS_PORT=6379 \
  -e ANSIBLE_TOWER_URL=https://tower.example.com \
  -e ANSIBLE_TOWER_USERNAME=admin \
  -e ANSIBLE_TOWER_PASSWORD=password \
  ansible-service
```

## Configuration

Le service utilise les variables d'environnement suivantes :

### Serveur
- `ANSIBLE_SERVICE_PORT` : Port du serveur (défaut: 8083)
- `ENV` : Environnement (development/production)
- `LOG_LEVEL` : Niveau de log (info/debug/warning/error)

### Redis
- `REDIS_HOST` : Host Redis (défaut: localhost)
- `REDIS_PORT` : Port Redis (défaut: 6379)
- `REDIS_PASSWORD` : Mot de passe Redis (optionnel)
- `REDIS_DB` : Base de données Redis (défaut: 0)
- `ANSIBLE_CACHE_TTL` : TTL du cache en secondes (défaut: 300)

### Ansible Tower
- `ANSIBLE_TOWER_URL` : URL de l'instance Ansible Tower (ex: https://tower.example.com)
- `ANSIBLE_TOWER_USERNAME` : Nom d'utilisateur pour l'authentification
- `ANSIBLE_TOWER_PASSWORD` : Mot de passe pour l'authentification
- `ANSIBLE_TOWER_VERIFY_SSL` : Vérifier les certificats SSL (défaut: true)

## API

### Documentation interactive

Une fois le service démarré, la documentation interactive est disponible :
- Swagger UI : http://localhost:8083/docs
- ReDoc : http://localhost:8083/redoc

### Santé

- `GET /health` - Vérifier l'état du service

### Jobs

- `GET /api/v1/ansible/jobs` - Lister les jobs (pagination)
  - Paramètres : `page`, `page_size`
- `GET /api/v1/ansible/jobs/{job_id}` - Détails d'un job
  - Paramètres : `include_stdout` (booléen, optionnel)
- `GET /api/v1/ansible/jobs/history` - Historique des jobs
  - Paramètres : `template_id` (optionnel), `page`, `page_size`

### Inventaires

- `GET /api/v1/ansible/inventories` - Lister les inventaires (pagination)
- `GET /api/v1/ansible/inventories/{inventory_id}` - Détails d'un inventaire
- `GET /api/v1/ansible/inventories/{inventory_id}/hosts` - Hôtes d'un inventaire

### Templates de jobs

- `GET /api/v1/ansible/job-templates` - Lister les templates (pagination)
- `GET /api/v1/ansible/job-templates/{template_id}` - Détails d'un template
- `POST /api/v1/ansible/job-templates/{template_id}/launch` - Lancer un job depuis un template
  - Body : `{"extra_vars": {...}}` (optionnel)

### Projets

- `GET /api/v1/ansible/projects` - Lister les projets (pagination)
- `GET /api/v1/ansible/projects/{project_id}/playbooks` - Playbooks d'un projet

## Exemples

### Lister les jobs

```bash
curl http://localhost:8083/api/v1/ansible/jobs?page=1&page_size=20
```

### Récupérer les détails d'un job avec sortie standard

```bash
curl "http://localhost:8083/api/v1/ansible/jobs/123?include_stdout=true"
```

### Lister les inventaires

```bash
curl http://localhost:8083/api/v1/ansible/inventories
```

### Lancer un template de job

```bash
curl -X POST http://localhost:8083/api/v1/ansible/job-templates/42/launch \
  -H "Content-Type: application/json" \
  -d '{
    "extra_vars": {
      "variable1": "value1",
      "variable2": "value2"
    }
  }'
```

### Récupérer l'historique des jobs

```bash
# Tous les jobs
curl http://localhost:8083/api/v1/ansible/jobs/history

# Jobs d'un template spécifique
curl "http://localhost:8083/api/v1/ansible/jobs/history?template_id=42"
```

## Structure du code

```
ansible-service/
├── main.py                          # Point d'entrée FastAPI
├── requirements.txt                 # Dépendances Python
├── Dockerfile                      # Image Docker
├── internal/
│   ├── config/                     # Configuration
│   │   └── config.py
│   ├── cache/                      # Cache Redis
│   │   └── redis.py
│   ├── client/                     # Client Ansible Tower
│   │   └── tower_client.py
│   ├── handler/                    # Handlers HTTP
│   │   └── ansible_handler.py
│   ├── models/                     # Modèles de données Pydantic
│   │   └── models.py
│   └── service/                    # Logique métier
│       └── ansible_service.py
└── README.md                       # Documentation
```

## Développement

### Installation des dépendances de développement

```bash
pip install -r requirements.txt
```

### Exécution en mode développement

```bash
# Avec rechargement automatique
uvicorn main:app --host 0.0.0.0 --port 8083 --reload
```

### Tests

```bash
# À venir : tests unitaires et d'intégration
pytest
```

## Notes importantes

- Le service peut fonctionner sans Ansible Tower configuré, mais les endpoints retourneront des erreurs 503
- Le cache Redis est optionnel - le service fonctionnera sans cache mais avec des performances réduites
- L'authentification utilise des tokens Ansible Tower (Token Authentication)
- Les réponses sont paginées par défaut (20 éléments par page)

## Fonctionnalités avancées

### Analyse approfondie des playbooks
- Parser YAML complet des playbooks
- Extraction des tâches, rôles, handlers, variables
- Statistiques détaillées (modules utilisés, hôtes ciblés, etc.)
- Endpoint: `POST /api/v1/ansible/playbooks/analyze`

### Webhooks
- Réception des événements Ansible Tower
- Traitement asynchrone des événements
- Endpoint: `POST /api/v1/webhooks/ansible-tower`

### Streaming temps réel
- WebSocket pour les logs de jobs en temps réel
- Mises à jour de statut en direct
- Endpoint: `WS /api/v1/ansible/jobs/{job_id}/stream`

### Gestion des credentials
- CRUD complet pour les credentials
- Endpoints: `GET/POST/PUT/DELETE /api/v1/ansible/credentials`

### Gestion des organisations
- CRUD complet pour les organisations
- Endpoints: `GET/POST/PUT/DELETE /api/v1/ansible/organizations`

### Métriques Prometheus
- Exposition des métriques au format Prometheus
- Compteurs, histogrammes et jauges
- Endpoint: `GET /metrics`

## Développement futur

- [ ] Intégration avec Kafka pour les événements asynchrones
- [ ] Tests unitaires et d'intégration complets
- [ ] Support de l'authentification OAuth2
- [ ] Support des inventaires dynamiques (smart inventories)
- [ ] Clonage Git pour analyser les playbooks directement depuis les projets
