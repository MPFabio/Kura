# Connecter GitHub Actions au module Pipeline (Kura)

Deux méthodes : **lier par API** (recommandé, sans URL publique) ou **webhooks** (temps réel).

---

## Méthode 1 : Lier depuis l'interface (recommandé)

Configurable directement dans Kura, sans modifier le code ni les variables d'environnement.

1. Ouvrez Kura → **Module Pipeline**
2. Cliquez sur **« Connecter un dépôt GitHub »**
3. Créez un [Personal Access Token](https://github.com/settings/tokens) (scope `repo` ou Actions read)
4. Saisissez le token et les dépôts (format `owner/repo`)
5. Cliquez sur **Enregistrer**, puis sur **Sync**

Synchronisation automatique toutes les 2 min. Compatible SaaS et self-hosted.

> **Alternative (self-hosted)** : vous pouvez aussi configurer `GITHUB_TOKEN` et `GITHUB_REPOS` dans le `.env` si vous déployez Kura vous-même.

---

## Méthode 2 : Webhooks (temps réel)

## Prérequis

1. **Pipeline-service accessible** : Kong doit router vers le pipeline-service et l'URL doit être **atteignable depuis Internet** (GitHub envoie les webhooks vers votre serveur).
2. **Redis** : Le pipeline-service stocke les runs dans Redis.

## Principe

- Le pipeline-service expose : `POST /api/v1/pipeline/webhooks/github`
- GitHub envoie des webhooks (`workflow_run`, `check_suite`) à cette URL
- Les runs sont stockés et visibles dans l'interface Kura

## Étape 1 : Obtenir l’URL du webhook

L’URL doit être **publique** pour que GitHub puisse l’atteindre.

| Environnement | URL du webhook |
|---------------|----------------|
| **Local (dev)** | Utiliser [ngrok](https://ngrok.com) ou [cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) pour exposer `http://localhost:8000` → ex. `https://abc123.ngrok.io/api/v1/pipeline/webhooks/github` |
| **GKE / Production** | `https://votre-domaine/api/v1/pipeline/webhooks/github` (via Kong / Ingress) |
| **Codespaces** | `https://${CODESPACE_NAME}-8000.${GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN}/api/v1/pipeline/webhooks/github` |

> **Important** : GitHub impose HTTPS pour les webhooks. En local, ngrok fournit une URL HTTPS automatiquement.

## Étape 2 : Configurer le webhook sur GitHub

1. Ouvrez votre dépôt GitHub (ex. `votre-org/ModulOps`)
2. **Settings** → **Webhooks** → **Add webhook**
3. Renseignez :
   - **Payload URL** : `https://votre-url/api/v1/pipeline/webhooks/github`
   - **Content type** : `application/json`
   - **Secret** (optionnel) : générez un secret pour vérifier les requêtes
   - **Which events** : cochez **"Let me select individual events"**
     - ✅ **Workflow runs** (principal pour GitHub Actions)
     - ✅ **Check runs** (optionnel, pour les check suites)
4. Cliquez sur **Add webhook**

## Étape 3 : Vérifier la connexion

1. Déclenchez un workflow GitHub Actions (push, PR, ou exécution manuelle)
2. Ouvrez Kura → **Module Pipeline**
3. Les runs devraient apparaître avec le statut (en cours, succès, échec)

## Dépannage

### Le webhook ne reçoit rien

- Vérifiez que l’URL est joignable depuis Internet
- Vérifiez les **Recent Deliveries** dans GitHub : Settings → Webhooks → votre webhook → voir les requêtes et les erreurs

### Erreur 502 / timeout

- Kong ou le pipeline-service ne sont peut‑être pas démarrés
- Test local : `curl -X POST http://localhost:8000/api/v1/pipeline/webhooks/github -H "Content-Type: application/json" -d '{}'` → doit répondre (même un JSON minimal)

### Les runs n’apparaissent pas dans Kura

- Vérifiez que Redis est bien démarré
- Vérifiez les logs du pipeline-service : `docker-compose logs pipeline-service`

## Exemple : test en local avec ngrok

```bash
# Terminal 1 : lancer ModulOps
docker-compose up -d

# Terminal 2 : exposer le port 8000
ngrok http 8000

# Copier l’URL HTTPS (ex. https://a1b2c3.ngrok-free.app)
# Webhook GitHub : https://a1b2c3.ngrok-free.app/api/v1/pipeline/webhooks/github
```

## Événements supportés

| Événement GitHub | Description |
|------------------|-------------|
| `workflow_run` | Exécutions de workflows GitHub Actions (démarré, terminé, etc.) |
| `check_suite` | Check suites (tests, validations) |
