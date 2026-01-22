# Changelog - Kura

Tous les changements notables de ce projet seront documentés dans ce fichier.

Le format est basé sur [Keep a Changelog](https://keepachangelog.com/fr/1.0.0/),
et ce projet adhère à [Semantic Versioning](https://semver.org/lang/fr/).

## [1.1.1] - 2026-01-22

### Corrigé
- **Frontend Terraform** :
  - Correction du bouton "Annuler" dans le dialog de modification de source cloud qui ne fermait pas correctement le dialog
  - Suppression du Dialog dupliqué qui causait des conflits
  - Amélioration de la fonction `handleCloseDialog` pour réinitialiser correctement tous les états

### Amélioré
- **Sécurité des credentials** :
  - Masquage des credentials existants en mode modification (affichage de `••••••••` au lieu de champs vides)
  - Empêchement de la copie des credentials masqués (événement `onCopy` bloqué)
  - Vidage automatique du masquage au focus pour permettre la saisie de nouvelles valeurs
  - Messages d'aide améliorés pour indiquer que les credentials existants sont masqués
  - Support du masquage pour tous les types de credentials : GCP JSON, AWS Access Key/Secret, Azure Account Key/Connection String

### Technique
- Réorganisation du code pour éviter les Dialogs dupliqués
- Amélioration de la gestion des états React pour la fermeture des dialogs

## [1.1.0] - 2026-01-20

### Ajouté
- **Module Terraform** : Service Terraform complet avec gestion des états
  - Upload et parsing de fichiers tfstate
  - Visualisation des ressources, outputs et métadonnées
  - Suppression d'états Terraform
  - Interface utilisateur avec onglets "États" et "Sources de synchronisation"
  
- **Synchronisation cloud** : Intégration avec les providers cloud
  - Support AWS S3 pour la synchronisation des états Terraform
  - Support Azure Blob Storage avec credentials (Account Key ou Connection String)
  - Support GCP Cloud Storage avec Service Account JSON
  - Synchronisation automatique configurable (intervalle personnalisable)
  - Création d'états depuis les sources cloud
  - Modification et suppression des sources de synchronisation
  - Interface de gestion complète des sources (ajout, modification, suppression, synchronisation manuelle)

- **Détection de drift** : Comparaison en temps réel avec l'infrastructure réelle
  - Détection automatique lors de la synchronisation depuis GCP
  - Support GCP Compute Engine :
    - Instances Compute Engine (status, machine type, zone, etc.)
    - Réseaux VPC (description, mode auto, etc.)
    - Firewalls (règles, ports, sources, cibles)
    - Adresses IP statiques (status, type, région)
  - Affichage détaillé des différences détectées (attributs, valeurs attendues vs réelles)
  - Dialog de résultats avec statuts (in_sync, drifted, missing, unknown)
  - Interface utilisateur pour déclencher la détection manuellement

- **Cache Redis amélioré** : Persistance des états Terraform
  - TTL de 30 jours pour les états Terraform (au lieu de 5 minutes)
  - Persistance des états entre redémarrages
  - Méthode `Keys` pour lister les états par pattern

### Modifié
- **Frontend Terraform** :
  - Ajout d'une colonne "#" pour la numérotation des ressources
  - Correction du bouton "Modifier" pour les sources de synchronisation
  - Amélioration de l'affichage des résultats de drift
  - Correction des erreurs de syntaxe JSX (caractères spéciaux dans les helperText)

- **Service Terraform** :
  - Intégration de la détection de drift dans le processus de synchronisation
  - Support des types complexes pour les outputs Terraform (interface{} au lieu de string)
  - Support des dépendances flexibles pour les ressources (interface{} au lieu de struct)
  - Génération automatique d'IDs uniques pour les états créés depuis les sources

### Technique
- Architecture de détection de drift extensible avec interfaces
- Intégration GCP Compute API (`cloud.google.com/go/compute`)
- Gestion sécurisée des credentials (chiffrement AES-GCM)
- Background jobs pour la synchronisation automatique

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

[1.1.0]: https://github.com/MPFabio/ModulOps/releases/tag/v1.1.0
[1.0.0]: https://github.com/MPFabio/ModulOps/releases/tag/v1.0.0
