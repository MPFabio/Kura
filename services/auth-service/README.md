# Auth Service - Service d'authentification Kura

Service d'authentification microservice développé en Go avec support JWT, gestion des utilisateurs et rôles.

## Fonctionnalités

- ✅ Authentification JWT avec tokens et refresh tokens
- ✅ Gestion des utilisateurs (création, mise à jour, suppression)
- ✅ Gestion des rôles (admin, user)
- ✅ Middleware d'authentification et d'autorisation
- ✅ Hachage sécurisé des mots de passe (bcrypt)
- ✅ Support OAuth2 (préparé pour Google, GitHub)
- ✅ Base de données PostgreSQL
- ✅ API REST complète

## Structure

```
auth-service/
├── main.go                    # Point d'entrée
├── go.mod                     # Dépendances Go
├── Dockerfile                 # Image Docker
├── internal/
│   ├── config/               # Configuration
│   ├── models/               # Modèles de données
│   ├── repository/           # Accès à la base de données
│   ├── service/              # Logique métier
│   └── handler/              # Handlers HTTP
└── README.md
```

## Configuration

Le service utilise les variables d'environnement suivantes :

- `AUTH_SERVICE_PORT` : Port du serveur (défaut: 8080)
- `POSTGRES_HOST` : Hôte PostgreSQL
- `POSTGRES_PORT` : Port PostgreSQL (défaut: 5432)
- `POSTGRES_USER` : Utilisateur PostgreSQL
- `POSTGRES_PASSWORD` : Mot de passe PostgreSQL
- `POSTGRES_DB` : Nom de la base de données
- `JWT_SECRET` : Secret pour signer les tokens JWT (obligatoire)
- `JWT_EXPIRATION` : Durée de validité des tokens (défaut: 24h)
- `ENV` : Environnement (development, production)
- `LOG_LEVEL` : Niveau de log (info, debug, error)

## API Endpoints

### Authentification

- `POST /api/v1/auth/register` - Inscription d'un nouvel utilisateur
- `POST /api/v1/auth/login` - Connexion
- `POST /api/v1/auth/refresh` - Rafraîchir le token
- `POST /api/v1/auth/logout` - Déconnexion
- `GET /api/v1/auth/me` - Récupérer l'utilisateur actuel (authentifié)
- `PUT /api/v1/auth/me` - Mettre à jour l'utilisateur actuel (authentifié)
- `PUT /api/v1/auth/password` - Changer le mot de passe (authentifié)

### Administration (nécessite le rôle admin)

- `GET /api/v1/admin/users` - Liste des utilisateurs
- `GET /api/v1/admin/users/:id` - Détails d'un utilisateur
- `PUT /api/v1/admin/users/:id` - Mettre à jour un utilisateur
- `DELETE /api/v1/admin/users/:id` - Supprimer un utilisateur
- `PUT /api/v1/admin/users/:id/role` - Mettre à jour les rôles d'un utilisateur

### Santé

- `GET /health` - Vérification de santé du service

## Exemples d'utilisation

### Inscription

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "username": "user",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Connexion

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Accès protégé

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <token>"
```

## Développement local

```bash
# Installer les dépendances
go mod download

# Démarrer le service (nécessite PostgreSQL en cours d'exécution)
go run main.go
```

## Build Docker

```bash
docker build -t kura-auth-service .
```

## Déploiement

Le service peut être déployé via Docker Compose ou Kubernetes. Voir la documentation principale du projet pour plus de détails.
