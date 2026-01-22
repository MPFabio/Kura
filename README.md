# Kura - Plateforme DevOps Microservices

[![Version](https://img.shields.io/badge/version-1.1.0-blue.svg)](VERSION)
[![GitHub release](https://img.shields.io/github/v/release/MPFabio/ModulOps?label=release)](https://github.com/MPFabio/Kura/releases)
[![GitHub tag](https://img.shields.io/github/v/tag/MPFabio/ModulOps?label=tag)](https://github.com/MPFabio/Kura/tags)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Plateforme DevOps unifiée pour la gestion de clusters Kubernetes, Terraform, Ansible et pipelines CI/CD.

## Architecture

Architecture microservices avec séparation des responsabilités par domaine :
- **API Gateway** : Kong pour le routage et l'authentification
- **Services Backend** : Go et Python pour les différents modules
- **Event Bus** : Kafka pour la communication asynchrone
- **Cache** : Redis pour les performances
- **Base de données** : PostgreSQL pour la persistance
- **Observabilité** : Prometheus et Grafana pour les métriques

## Prérequis

### Pour le développement local (Docker Compose)

- Docker 20.10+
- Docker Compose 2.0+ (ou docker compose)

### Pour Kubernetes

- kubectl 1.28+
- Un cluster Kubernetes local :
  - [minikube](https://minikube.sigs.k8s.io/docs/start/) (recommandé)
  - [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)

## Installation

### Option 1 : Développement local avec Docker Compose

```bash
# Démarrer tous les services
./scripts/setup-local.sh

# Ou manuellement
docker-compose up -d
```

Les services seront disponibles sur :

**Services accessibles via navigateur web :**
- Kong Gateway: http://localhost:8000
- Kong Admin: http://localhost:8001
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

**Services nécessitant des clients spécifiques :**
- PostgreSQL: `localhost:5432` (utiliser `psql` ou un client SQL)
- Redis: `localhost:6379` (utiliser `redis-cli` ou un client Redis)
- Kafka: `localhost:9092` (utiliser un client Kafka)

> **Note importante** : PostgreSQL, Redis et Kafka ne sont **pas** des serveurs web et ne peuvent pas être accédés via un navigateur. Consultez `docs/ACCES_SERVICES.md` pour plus de détails sur comment accéder à ces services.

### Option 2 : Déploiement sur Kubernetes

```bash
# Configurer et déployer l'infrastructure
./scripts/setup-k8s.sh

# Ou manuellement avec kubectl
kubectl apply -k infrastructure/k8s/
```

Pour accéder aux services, utilisez port-forward :

```bash
# PostgreSQL
kubectl port-forward svc/postgres 5432:5432 -n kura

# Redis
kubectl port-forward svc/redis 6379:6379 -n kura

# Kafka
kubectl port-forward svc/kafka 9092:9092 -n kura

# Kong Gateway
kubectl port-forward svc/kong 8000:8000 -n kura

# Prometheus
kubectl port-forward svc/prometheus 9090:9090 -n kura

# Grafana
kubectl port-forward svc/grafana 3000:3000 -n kura
```

## Structure du projet

```
Kura/
├── services/              # Services microservices
│   ├── api-gateway/       # Kong configuration
│   ├── auth-service/      # Service d'authentification (Go)
│   ├── k8s-service/       # Service Kubernetes (Go)
│   ├── terraform-service/ # Service Terraform (Go)
│   ├── ansible-service/   # Service Ansible Tower (Python)
│   ├── pipeline-service/  # Service Pipelines (Go)
│   ├── alert-service/     # Service d'alertes (Go)
│   └── metrics-service/   # Service de métriques (Go)
├── frontend/              # Application React + TypeScript
├── infrastructure/        # Configuration infrastructure
│   ├── k8s/              # Manifests Kubernetes
│   ├── docker/           # Dockerfiles
│   └── helm/             # Charts Helm
├── shared/                # Bibliothèques partagées
│   ├── go-common/        # Utilitaires Go communs
│   └── proto/            # Définitions gRPC
├── scripts/               # Scripts de déploiement
└── docs/                  # Documentation
```

## Services déployés

### Infrastructure de base

- **PostgreSQL 15** : Base de données principale
- **Redis 7** : Cache distribué avec persistance (AOF + RDB)
- **Kafka + Zookeeper** : Message broker pour les événements
- **Kong 3.4** : API Gateway avec routage et authentification
- **Prometheus** : Collecte de métriques
- **Grafana** : Visualisation des métriques et dashboards

### Services applicatifs

- **Auth Service** : Authentification centralisée avec JWT et refresh tokens
- **K8s Service** : Gestion des clusters Kubernetes avec terminal interactif
- **Terraform Service** : Gestion des états Terraform avec synchronisation cloud et détection de drift
  - Synchronisation automatique depuis S3, Azure Blob Storage, GCP Cloud Storage
  - Détection de drift en temps réel via APIs cloud (GCP, AWS, Azure)
  - Parsing et visualisation des états Terraform
  - Gestion des sources de synchronisation avec modification et suppression

## Développement

### Arrêter les services locaux

```bash
docker-compose down
```

### Voir les logs

```bash
docker-compose logs -f [service-name]
```

### Vérifier le statut sur Kubernetes

```bash
kubectl get all -n kura
```

## Fonctionnalités principales

### Module Terraform (v1.1.0)

- **Gestion des états Terraform** : Upload, visualisation et suppression d'états
- **Synchronisation cloud** : 
  - Support S3, Azure Blob Storage, GCP Cloud Storage
  - Synchronisation automatique configurable
  - Modification et suppression des sources de synchronisation
- **Détection de drift** :
  - Comparaison en temps réel avec l'infrastructure réelle via APIs cloud
  - Support GCP Compute Engine (instances, réseaux, firewalls, adresses IP)
  - Affichage détaillé des différences détectées
- **Interface utilisateur** :
  - Onglets "États" et "Sources de synchronisation"
  - Visualisation des ressources avec numérotation
  - Dialog de résultats de drift avec détails

### Module Kubernetes

- Gestion complète des clusters (namespaces, pods, deployments, services, etc.)
- Terminal interactif via WebSocket
- Actions en masse (bulk actions)
- Recherche et filtrage avancés

### Authentification

- Inscription et connexion utilisateurs
- JWT avec refresh tokens
- Gestion des rôles (admin, user)

## Documentation

Voir `docs/` pour plus de détails sur l'architecture et les services.

## Licence

[À définir]
