# 🚀 Guide de démarrage rapide - Auth Service

## 📖 Ce qui a été créé

J'ai créé un **service d'authentification complet** en Go qui permet de :

1. **Créer des comptes utilisateurs** (inscription)
2. **Se connecter** et obtenir un token JWT
3. **Protéger des routes** avec authentification
4. **Gérer les utilisateurs** (modifier, supprimer)
5. **Gérer les rôles** (admin, user)

Le service stocke tout dans **PostgreSQL** et utilise **JWT** pour l'authentification.

## ⚡ Démarrage rapide (3 étapes)

### Étape 1 : Démarrer PostgreSQL

```bash
# Depuis la racine du projet ModulOps
docker-compose up -d postgres

# Attendre 10 secondes que PostgreSQL soit prêt
sleep 10
```

### Étape 2 : Démarrer le service d'authentification

```bash
# Toujours depuis la racine du projet
docker-compose up -d auth-service

# Vérifier que ça fonctionne
docker-compose logs auth-service
```

Vous devriez voir : `Service d'authentification démarré sur le port 8080`

### Étape 3 : Tester

**Option A : Script automatique (recommandé)**

```bash
cd services/auth-service
bash test.sh
```

**Option B : Tests manuels**

```bash
# 1. Vérifier que le service répond
curl http://localhost:8080/health

# 2. Créer un utilisateur
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","username":"test","password":"password123","first_name":"Test","last_name":"User"}'

# 3. Se connecter (copiez le token de la réponse)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# 4. Utiliser le token (remplacez VOTRE_TOKEN)
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer VOTRE_TOKEN"
```

## ✅ Vérifications

### Le service fonctionne si :

1. ✅ `curl http://localhost:8080/health` retourne `{"status":"ok"}`
2. ✅ Vous pouvez créer un utilisateur avec `register`
3. ✅ Vous pouvez vous connecter avec `login` et obtenir un token
4. ✅ Vous pouvez accéder à `/auth/me` avec le token

### Si ça ne fonctionne pas :

1. **PostgreSQL n'est pas démarré** :
   ```bash
   docker-compose ps postgres
   # Si pas de résultat, lancez : docker-compose up -d postgres
   ```

2. **Le service ne démarre pas** :
   ```bash
   docker-compose logs auth-service
   # Vérifiez les erreurs dans les logs
   ```

3. **Erreur de connexion à la base** :
   - Vérifiez que PostgreSQL est bien démarré
   - Attendez 10-15 secondes après le démarrage de PostgreSQL

## 📁 Structure créée

```
services/auth-service/
├── main.go                    # Point d'entrée du service
├── go.mod                     # Dépendances Go
├── Dockerfile                 # Pour créer l'image Docker
├── internal/
│   ├── config/               # Configuration (variables d'env)
│   ├── models/              # Modèles de données (User, etc.)
│   ├── repository/          # Accès à PostgreSQL
│   ├── service/             # Logique métier (auth, users)
│   └── handler/             # Routes HTTP (API REST)
├── TEST.md                   # Guide de test détaillé
├── test.sh                   # Script de test automatisé
└── README.md                 # Documentation
```

## 🔍 Comment ça marche ?

1. **Vous créez un utilisateur** → Stocké dans PostgreSQL avec mot de passe hashé
2. **Vous vous connectez** → Le service génère un JWT token
3. **Vous utilisez le token** → Le service vérifie que vous êtes authentifié
4. **Vous accédez aux routes protégées** → Le middleware vérifie le token

## 📚 Documentation complète

Pour plus de détails, consultez :
- `TEST.md` : Guide de test complet avec tous les endpoints
- `README.md` : Documentation technique du service

## 🎯 Prochaines étapes

Une fois que tout fonctionne :
1. Intégrer avec Kong API Gateway
2. Créer un utilisateur admin
3. Connecter le frontend
4. Ajouter d'autres services qui utiliseront l'auth
