## Architecture globale ModulOps

Ce document décrit l’architecture haut niveau de la plateforme ModulOps ainsi que le détail du service d’authentification (`auth-service`).

### Vue globale

```mermaid
graph TD
    FE[Frontend React+TS]
    KONG[Kong API Gateway]

    subgraph Services
        AUTH[Auth Service (Go)]
        K8S[K8s Service (Go)]
        TF[Terraform Service (Go)]
        ANS[Ansible Service (Python)]
        PIPE[Pipeline Service (Go)]
        ALERT[Alert Service (Go)]
        METRICS[Metrics Service (Go)]
    end

    subgraph Infra
        PG[(PostgreSQL)]
        REDIS[(Redis)]
        KAFKA[(Kafka)]
        PROM[(Prometheus)]
        GRAF[Grafana]
    end

    FE --> KONG

    KONG --> AUTH
    KONG --> K8S
    KONG --> TF
    KONG --> ANS
    KONG --> PIPE
    KONG --> ALERT
    KONG --> METRICS

    K8S --> KAFKA
    TF --> KAFKA
    ANS --> KAFKA
    PIPE --> KAFKA

    K8S --> REDIS
    TF --> REDIS
    ANS --> REDIS

    K8S --> PG
    TF --> PG
    ANS --> PG

    KAFKA --> ALERT
    KAFKA --> METRICS

    METRICS --> PROM
    PROM --> GRAF
```

## Architecture détaillée du service d’authentification

Cette vue se concentre uniquement sur `auth-service` et ses dépendances directes.

```mermaid
graph TD
    FE[Client / Frontend / Autres services]
    KONG[Kong API Gateway]
    AUTH[Auth Service (Go)]

    subgraph Auth_Service
        H[Handlers HTTP (Gin)]
        S[AuthService<br/>Logique métier]
        R[Repository<br/>Accès PostgreSQL]
        JWT[JWT & Refresh Tokens]
    end

    PG[(PostgreSQL)]

    FE -->|"HTTP /api/v1/auth/*"| KONG
    KONG --> AUTH

    AUTH --> H
    H --> S
    S --> R
    S --> JWT
    R --> PG
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

