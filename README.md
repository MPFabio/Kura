# Kura

[![Version](https://img.shields.io/badge/version-1.3.0-blue.svg)](VERSION)
[![GitHub release](https://img.shields.io/github/v/release/MPFabio/Kura?label=release)](https://github.com/MPFabio/Kura/releases)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Console d'opération DevOps unifiée — Kubernetes, ArgoCD, Terraform, Ansible, Pipelines CI/CD, Registre OCI, OpenBao et Métriques dans une seule interface.

## Architecture

```
Frontend React/TS → Kong API Gateway → auth-service
                                     → k8s-service
                                     → terraform-service
                                     → ansible-service
                                     → pipeline-service
                                     → metrics-service
                                     → code-service
                                     → vault-service

Redis  — cache distribué
PostgreSQL — persistance
Prometheus/VictoriaMetrics + Grafana + Loki + Tempo — observabilité
```

Voir [docs/architecture.md](docs/architecture.md) pour les diagrammes détaillés.

## Prérequis

- Docker 20.10+
- Docker Compose 2.0+

## Démarrage rapide

```bash
cp .env.example .env
# Remplir les variables dans .env

docker compose -p kura --env-file .env -f docker-compose.yml up -d
```

## Déploiement production

L'infrastructure et le déploiement sont gérés dans le repo [Kuro](https://codeberg.org/MPFabio/Kuro) (Terraform + cloud-init). Sur la VM, le démarrage clone ce repo et lance `docker compose up -d`.

## Services

| Service | Port | Description |
|---------|------|-------------|
| Frontend | 5173 | Interface React/TypeScript |
| Kong Gateway | 8000 | API Gateway (point d'entrée unique) |
| Auth Service | 8080 | Authentification JWT |
| K8s Service | 8081 | Gestion clusters Kubernetes, ArgoCD, registre Zot |
| Terraform Service | 8082 | États Terraform + détection de drift |
| Ansible Service | 8083 | Intégration Semaphore |
| Pipeline Service | 8084 | Pipelines CI/CD (Forgejo Actions) |
| Metrics Service | 8086 | Agrégation Prometheus/Grafana |
| Vault Service | 8087 | Intégration OpenBao |
| Code Service | 8088 | Intégration Forgejo/Codeberg, dépôts GitOps |
| Grafana | 3000 | Dashboards métriques |
| VictoriaMetrics | 9090 | Collecte métriques |
| Semaphore | 3001 | UI Ansible |

## Fonctionnalités

### Kubernetes
- Gestion namespaces, pods, deployments, services
- Terminal interactif via WebSocket
- Actions en masse, recherche et filtrage

### Terraform
- Upload et visualisation d'états
- Synchronisation depuis S3, Azure Blob Storage, GCP Cloud Storage
- Détection de drift en temps réel via APIs cloud (GCP, AWS, Azure)

### Ansible
- Intégration Semaphore (UI incluse sur `/ansible`)
- Exécution de playbooks, inventaires

### ArgoCD & Catalogue Helm
- Installation d'ArgoCD sur le cluster actif, auto-gestion via le pattern « app of apps »
- Création d'Applications depuis un dépôt Git ou le catalogue Helm (ArtifactHub)
- Suivi de synchronisation/santé, historique et rollback

### Registre (Zot)
- Registre OCI privé pour images de conteneurs et charts Helm
- Vérification de signature Cosign

### Code (Forgejo/Codeberg)
- Gestion des dépôts et branches GitOps utilisés par ArgoCD
- Commit des manifests depuis l'interface Kura

### OpenBao
- Connexion à une instance OpenBao existante
- Consultation et gestion des secrets

### Pipelines
- Synchronisation avec Forgejo Actions
- Suivi des runs CI/CD

### Observabilité
- Dashboards Grafana accessibles sur `/grafana`
- Métriques VictoriaMetrics, logs Loki, traces Tempo

## Structure

```
Kura/
├── services/
│   ├── auth-service/       # Go — Authentification JWT
│   ├── k8s-service/        # Go — Kubernetes, ArgoCD, registre Zot
│   ├── terraform-service/  # Go — Terraform + drift
│   ├── ansible-service/    # Python — Ansible/Semaphore
│   ├── pipeline-service/   # Go — CI/CD
│   ├── metrics-service/    # Go — Métriques
│   ├── code-service/       # Go — Forgejo/Codeberg, GitOps
│   └── vault-service/      # Go — OpenBao
├── frontend/               # React + TypeScript + Vite
├── infrastructure/
│   ├── docker/             # Kong, Caddy, Prometheus config
│   └── k8s/                # Manifests Kubernetes
├── docs/                   # Architecture, guides
└── scripts/                # Scripts utilitaires
```

## Commandes utiles

```bash
# Logs d'un service
docker compose -p kura logs -f auth-service

# Statut des conteneurs
docker compose -p kura ps

# Arrêter la stack
docker compose -p kura down

# Rebuild d'un service
docker compose -p kura build --parallel && docker compose -p kura up -d
```

## Licence

MIT
