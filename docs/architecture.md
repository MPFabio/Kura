## Contexte et objectifs du projet

Kura est une plateforme DevOps unifiée qui vise à donner **une vue centrale** sur plusieurs briques déjà très utilisées en entreprise : Kubernetes, Terraform, Ansible, pipelines CI/CD, métriques et alertes.

### Ce qui existe déjà sur le marché

- **Outils spécialisés mais isolés** :
  - Portails Kubernetes (Lens, Octant, dashboards maison),
  - Outils Terraform (Terraform Cloud, Atlantis),
  - Interfaces Ansible Tower / AWX,
  - Outils CI/CD (GitHub Actions, GitLab CI, Jenkins),
  - Stacks d’observabilité (Prometheus / Grafana).
  
- **Portails développeur / Developer Portals** :
  - **Backstage (Spotify)** : plateforme open-source pour créer des portails développeur, catalogue de services, documentation, plugins CI/CD.
  - **Port** : plateforme SaaS similaire à Backstage, focus sur la découverte et la gouvernance.
  - **Cortex** : catalogue de microservices avec scoring et dépendances.
  
- **Problèmes communs à ces solutions** :
  - Les outils spécialisés isolés : les équipes jonglent entre **plusieurs interfaces**, difficile d’avoir une **vision transversale** (ex. : « ce pipeline déploie quels clusters / quelles ressources Terraform ? »), l’authentification et les rôles sont souvent **dupliqués** dans chaque outil.
  - Les Developer Portals (Backstage, Port, Cortex) : excellents pour la **découverte** et la **documentation**, mais souvent **légers sur l’opérationnel** (gestion active des clusters K8s, exécution Terraform, jobs Ansible, alertes temps réel). Ils sont plutôt orientés « catalogue » que « console d’opération ».

### La valeur ajoutée de Kura

- **Point d’entrée unique** pour les équipes Ops / DevOps : un seul portail, une seule API Gateway.
- **Agrégation** des infos clés (clusters, états Terraform, jobs Ansible, pipelines, métriques) au même endroit.
- **Modèle d’authentification centralisé** (auth-service) et rôles homogènes sur tous les modules.
- **Événements corrélés** via Kafka (ex. : un déploiement Terraform qui déclenche des métriques et des alertes associées).
- **Différenciation vs Backstage/Port** :
  - Kura se concentre sur l’**opérationnel actif** (exécution Terraform, gestion K8s, jobs Ansible) plutôt que sur le catalogue/documentation.
  - Architecture **microservices native** avec bus d’événements Kafka pour corréler les actions entre systèmes.
  - **Focus Ops/DevOps** : console d’opération plutôt que portail développeur (même si les deux peuvent coexister).

### Pourquoi c’est faisable

- Tous les composants ciblés exposent déjà :
  - Des **API HTTP** (Kubernetes, Terraform Cloud, Ansible Tower, outils CI/CD),
  - Des **endpoints d’observabilité** (Prometheus, logs, etc.).
- L’architecture choisie (microservices + Kafka + Postgres + Redis) est **classique et éprouvée** dans l’écosystème cloud-native.
- Le périmètre est découpé par domaines (`auth-service`, `k8s-service`, `terraform-service`, etc.) ce qui permet d’avancer **service par service** sans tout faire d’un coup.

### Pourquoi c’est utile

- Réduit la **charge cognitive** : moins d’outils à connaître pour les équipes.
- Facilite les **diagnostics cross-systèmes** (pipeline → déploiement → métriques → alertes).
- Permet de mettre en place des **garde-fous globaux** (rôles, droits, observabilité, audit) au niveau de la plateforme, pas outil par outil.

---

## Architecture globale Kura

Ce document décrit l’architecture haut niveau de la plateforme Kura ainsi que le détail du service d’authentification (`auth-service`).

### Vue globale

```mermaid
graph TB
    Frontend[Frontend React+TS] --> Gateway[API Gateway Kong]
    Gateway --> Auth[Auth Service Go]
    Gateway --> K8s[K8s Service Go]
    Gateway --> Terraform[Terraform Service Go]
    Gateway --> Ansible[Ansible Service Python]
    Gateway --> Pipeline[Pipeline Service Go]
    Gateway --> Vault[Vault Service Go - OpenBao]
    
    K8s --> Kafka[Kafka Event Bus]
    Terraform --> Kafka
    Ansible --> Kafka
    Pipeline --> Kafka
    
    Kafka --> Alert[Alert Service Go]
    Kafka --> Metrics[Metrics Service Go]
    
    K8s --> Redis[(Redis Cache)]
    Terraform --> Redis
    Ansible --> Redis
    Vault --> Redis
    
    K8s --> Postgres[(PostgreSQL)]
    Terraform --> Postgres
    Ansible --> Postgres
    
    Terraform --> S3[(AWS S3)]
    Terraform --> Azure[(Azure Blob)]
    Terraform --> GCP[(GCP Storage)]
    
    Terraform --> GCPAPI[GCP Compute API]
    
    Vault --> ExtVault[(OpenBao/Vault du client)]
    
    Metrics --> Prometheus[(Prometheus)]
    Metrics --> Grafana[Grafana]
```

## Bus d’événements Kafka — Flux implémentés

Kafka est le bus d’événements central de Kura. Il découple les producteurs (services qui détectent des changements) des consommateurs (services qui réagissent à ces changements).

### Topics actifs

| Topic | Producteur | Description | Implémenté |
|-------|-----------|-------------|:----------:|
| `terraform.drift.detected` | `terraform-service` (drift_worker) | Émis quand un drift est détecté entre l’état Terraform stocké et la réalité cloud | ✅ |
| `pipeline.run.updated` | `pipeline-service` | Émis quand un run CI/CD change de statut (via webhook GitHub/GitLab) | 🔜 |
| `k8s.deployment.changed` | `k8s-service` | Émis quand un déploiement Kubernetes change d’état | 🔜 |

### Flux : Détection de drift Terraform

```mermaid
sequenceDiagram
    participant DW as DriftWorker (terraform-service)
    participant SS as SyncService
    participant TF as TerraformService
    participant K as Kafka (terraform.drift.detected)
    participant AS as AlertService (futur)

    DW->>SS: ListSources() — sources avec auto_sync=true
    SS-->>DW: [source1, source2, ...]
    DW->>TF: DetectDrift(ctx, stateFileID, credentials, provider)
    TF-->>DW: [DriftResult{status="modified"}, ...]
    Note over DW: drifted = true → emitDriftEvent()
    DW->>K: WriteMessages(DriftEventPayload{<br/>event_type: "terraform.drift.detected",<br/>state_file_id, source_id,<br/>drift_count, results})
    K-->>AS: Consume → déclencher alerte (futur)
```

### Structure du payload `terraform.drift.detected`

```json
{
  "event_type": "terraform.drift.detected",
  "state_file_id": "abc123",
  "source_id": "source-gcp-prod",
  "detected_at": "2026-06-03T14:32:00Z",
  "drift_count": 3,
  "results": [
    { "resource_address": "google_compute_instance.web", "status": "modified" },
    { "resource_address": "google_storage_bucket.assets", "status": "deleted" }
  ]
}
```

La **clé du message Kafka** est le `state_file_id`, ce qui garantit que les événements pour un même état Terraform sont traités dans l’ordre par le même consommateur (ordonnancement de partition).

### Graceful shutdown

Le `DriftWorker` démarre avec un `context.WithCancel`. Lors d’un `SIGTERM`, `Stop()` annule le contexte et attend que la goroutine `run()` termine proprement (via le canal `done`). Cela évite de laisser une détection de drift en cours être interrompue à mi-chemin.

Le `pipeline-service` utilise le même pattern : un `context.WithCancel` racine est propagé à la goroutine de sync GitHub. À l’arrêt, `rootCancel()` est appelé et `main` attend la fermeture de `syncDone` avant de continuer le shutdown HTTP.

---

## Architecture détaillée du service d’authentification

Cette vue se concentre uniquement sur `auth-service` et ses dépendances directes.

```mermaid
graph TD
    Client[Client Frontend Autres services] --> Kong[Kong API Gateway]
    Kong --> Auth[Auth Service Go]
    
    Auth --> Handlers[Handlers HTTP Gin]
    Handlers --> Service[AuthService Logique métier]
    Service --> Repository[Repository Accès PostgreSQL]
    Service --> JWT[JWT Refresh Tokens]
    Repository --> Postgres[(PostgreSQL)]
```

### Flux principal (MVP actuel)

- **Inscription** (`/api/v1/auth/register`)  
  - Le client envoie les infos → Kong → `auth-service`  
  - `Handlers` → `AuthService` → `Repository` → table `users` dans PostgreSQL.

- **Connexion** (`/api/v1/auth/login`)  
  - Vérification mot de passe (bcrypt)  
  - Génération d’un **JWT** + **refresh token**  
  - Persisté dans la table `refresh_tokens`.

- **Accès protégé** (`/api/v1/auth/me`, etc.)  
  - Le client envoie `Authorization: Bearer <JWT>`  
  - Middleware de `auth-service` valide le token et ajoute l’ID utilisateur au contexte.

En résumé :

- le schéma global explique **comment les services discutent entre eux** (Kong, Kafka, Postgres, Redis, Prometheus, Grafana) ;
- le schéma détaillé d’`auth-service` montre **comment l’authentification est centralisée et découplée** :
  - les `Handlers` gèrent uniquement HTTP,
  - `AuthService` porte la logique métier (mots de passe, rôles, JWT, refresh tokens),
  - le `Repository` encapsule l’accès à PostgreSQL.

