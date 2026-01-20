# Guide de test - Auth Service

Ce guide vous explique comment tester le service d'authentification pour vérifier que tout fonctionne correctement.

##Ce qui a été créé

Le service d'authentification est un microservice Go qui fournit :

1. **Authentification JWT** : Connexion avec génération de tokens
2. **Gestion des utilisateurs** : Création, modification, suppression
3. **Gestion des rôles** : Admin et User
4. **API REST** : Endpoints HTTP pour toutes les opérations
5. **Base de données** : Stockage dans PostgreSQL avec migrations automatiques

##Étape 1 : Démarrer les services nécessaires

### Option A : Avec Docker Compose (recommandé)

```bash
# Démarrer PostgreSQL (nécessaire pour le service d'auth)
docker-compose up -d postgres

# Attendre que PostgreSQL soit prêt (environ 10 secondes)
docker-compose ps postgres
```

### Option B : Vérifier que PostgreSQL tourne déjà

```bash
# Vérifier si PostgreSQL est déjà en cours d'exécution
docker ps | grep postgres
```

##Étape 2 : Démarrer le service d'authentification

### Option A : Avec Docker Compose

```bash
# Démarrer le service d'authentification
docker-compose up -d auth-service

# Voir les logs pour vérifier qu'il démarre correctement
docker-compose logs -f auth-service
```

Vous devriez voir :
```
Service d'authentification démarré sur le port 8080
```

### Option B : En local (pour développement)

```bash
# Aller dans le dossier du service
cd services/auth-service

# Installer les dépendances
go mod download

# Définir les variables d'environnement
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=modulops
export POSTGRES_PASSWORD=modulops
export POSTGRES_DB=modulops
export JWT_SECRET=test-secret-key-12345
export JWT_EXPIRATION=24h
export AUTH_SERVICE_PORT=8080
export ENV=development

# Lancer le service
go run main.go
```

##Étape 3 : Vérifier que le service fonctionne

### Test 1 : Vérifier la santé du service

```bash
curl http://localhost:8080/health
```

**Résultat attendu :**
```json
{
  "status": "ok",
  "service": "auth-service"
}
```

✅ Si vous voyez ce résultat, le service fonctionne !

## 🧪 Étape 4 : Tester les fonctionnalités

### Test 2 : Créer un utilisateur (Inscription)

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

**Résultat attendu :**
```json
{
  "message": "Utilisateur créé avec succès",
  "user": {
    "id": "uuid-here",
    "email": "test@example.com",
    "username": "testuser",
    "roles": ["user"],
    "first_name": "Test",
    "last_name": "User",
    "active": true,
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```

Si vous voyez ce résultat, l'inscription fonctionne !

### Test 3 : Se connecter (Login)

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Résultat attendu :**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid-here",
    "email": "test@example.com",
    "username": "testuser",
    "roles": ["user"],
    ...
  },
  "expires_at": "2024-01-02T12:00:00Z"
}
```

**Important** : Copiez le `token` pour les tests suivants !

### Test 4 : Accéder à une route protégée (Récupérer mon profil)

Remplacez `VOTRE_TOKEN` par le token obtenu à l'étape précédente :

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer VOTRE_TOKEN"
```

**Résultat attendu :**
```json
{
  "id": "uuid-here",
  "email": "test@example.com",
  "username": "testuser",
  "roles": ["user"],
  ...
}
```

✅ Si vous voyez vos informations, l'authentification JWT fonctionne !

### Test 5 : Rafraîchir le token

Remplacez `VOTRE_REFRESH_TOKEN` par le refresh_token obtenu au login :

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "VOTRE_REFRESH_TOKEN"
  }'
```

**Résultat attendu :** Un nouveau token et un nouveau refresh_token

✅ Si vous obtenez de nouveaux tokens, le refresh fonctionne !

### Test 6 : Modifier mon profil

```bash
curl -X PUT http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer VOTRE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Nouveau",
    "last_name": "Nom"
  }'
```

✅ Si vous voyez vos informations mises à jour, la modification fonctionne !

## 🔍 Vérifications supplémentaires

### Vérifier la base de données

```bash
# Se connecter à PostgreSQL
docker exec -it modulops-postgres psql -U modulops -d modulops

# Voir les utilisateurs créés
SELECT id, email, username, roles, created_at FROM users;

# Voir les refresh tokens
SELECT id, user_id, expires_at, revoked FROM refresh_tokens;

# Quitter
\q
```

### Vérifier les logs du service

```bash
# Voir les logs en temps réel
docker-compose logs -f auth-service

# Ou pour un conteneur spécifique
docker logs -f modulops-auth-service
```

## ❌ Résolution des problèmes

### Erreur : "connection refused" ou "cannot connect to database"

**Solution :**
1. Vérifiez que PostgreSQL est démarré : `docker-compose ps postgres`
2. Vérifiez les variables d'environnement (POSTGRES_HOST, etc.)
3. Attendez quelques secondes que PostgreSQL soit complètement prêt

### Erreur : "JWT_SECRET doit être défini"

**Solution :**
Définissez la variable d'environnement `JWT_SECRET` :
```bash
export JWT_SECRET=mon-secret-super-securise
```

### Erreur : "utilisateur non trouvé" lors du login

**Solution :**
Assurez-vous d'avoir créé un utilisateur avec `register` avant de faire `login`

### Erreur : "token invalide ou expiré"

**Solution :**
1. Vérifiez que vous utilisez le bon format : `Authorization: Bearer VOTRE_TOKEN`
2. Le token peut être expiré, refaites un login ou refresh

## 📊 Checklist de validation

- [ ] Le service démarre sans erreur
- [ ] `/health` retourne `{"status": "ok"}`
- [ ] Je peux créer un utilisateur avec `register`
- [ ] Je peux me connecter avec `login` et obtenir un token
- [ ] Je peux accéder à `/auth/me` avec le token
- [ ] Je peux rafraîchir mon token avec `refresh`
- [ ] Je peux modifier mon profil avec `PUT /auth/me`
- [ ] Les données sont bien stockées dans PostgreSQL

Si tous ces points sont validés, **le service fonctionne correctement !** ✅

## 🎯 Prochaines étapes

Une fois que tout fonctionne, vous pouvez :
1. Intégrer ce service avec Kong API Gateway
2. Créer un utilisateur admin pour tester les routes d'administration
3. Connecter le frontend à ce service
4. Ajouter d'autres services qui utiliseront ce service d'authentification
