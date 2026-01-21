# Changelog - Kura

Tous les changements notables de ce projet seront documentés dans ce fichier.

Le format est basé sur [Keep a Changelog](https://keepachangelog.com/fr/1.0.0/),
et ce projet adhère à [Semantic Versioning](https://semver.org/lang/fr/).

## [1.0.0] - 2026-01-20

### Ajouté
- **Authentification** : Service d'authentification complet avec JWT et refresh tokens
  - Inscription et connexion utilisateurs
  - Gestion des rôles (admin, user)
  - API REST complète (`/api/v1/auth/*`)

- **Module Kubernetes** : Interface complète de gestion Kubernetes
  - Liste et visualisation des namespaces, pods, deployments, services, configmaps, secrets, nodes
  - Recherche et filtrage des ressources
  - Actions sur les ressources : scale, delete
  - **Actions en masse (Bulk Actions)** : Sélection multiple et actions groupées
    - Suppression en masse de pods, deployments, services
    - Redémarrage en masse de pods
    - Scale en masse de deployments
  - Détails des ressources : YAML, logs, events
  - **Terminal interactif** : Exécution de commandes dans les pods via WebSocket
    - Support xterm.js pour une expérience terminal complète
    - Détection automatique des shells disponibles
    - Messages d'erreur informatifs pour les conteneurs sans shell

- **Frontend** : Interface utilisateur moderne
  - Thème sombre "Abyssal Glow" avec effets de lueur
  - Design responsive avec Material-UI
  - Navigation intuitive avec sidebar
  - Intégration React Query pour la gestion des données
  - WebSocket pour les mises à jour en temps réel

- **Infrastructure** : Stack complète
  - Docker Compose pour le développement local
  - Manifests Kubernetes pour le déploiement
  - Kong API Gateway pour le routage
  - PostgreSQL, Redis, Kafka, Zookeeper
  - Prometheus et Grafana pour l'observabilité

### Technique
- Architecture microservices avec séparation des responsabilités
- API Gateway centralisé (Kong)
- Event Bus (Kafka) pour la communication asynchrone
- Cache distribué (Redis)
- Authentification centralisée avec JWT

### Documentation
- README.md avec instructions d'installation
- Documentation d'architecture
- Guide d'accès aux services

[1.0.0]: https://github.com/MPFabio/ModulOps/releases/tag/v1.0.0
