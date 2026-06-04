# Connecter GitHub Actions au module Pipeline

Ce guide explique comment afficher vos exécutions GitHub Actions dans le module Pipelines de Kura.

## Option recommandée : Webhooks (temps réel)

Les webhooks permettent une mise à jour **immédiate** des exécutions dès qu’un workflow GitHub se termine. Aucun polling, aucune latence.

### 1. Configurer le webhook sur GitHub

1. Dans votre dépôt GitHub : **Settings** → **Webhooks** → **Add webhook**
2. **Payload URL** : saisissez l’URL de votre instance Kura :
   ```
   https://votre-domaine/api/v1/pipeline/webhooks/github
   ```
   (en local : `http://localhost:8000/api/v1/pipeline/webhooks/github`)
3. **Content type** : `application/json`
4. **Secret** (optionnel) : définissez un secret et configurez `GITHUB_WEBHOOK_SECRET` côté Kura
5. **Events** : sélectionnez « Let me select individual events » puis cochez :
   - `Workflow run` (GitHub Actions)
   - `Check suite` (alternative)
6. Cliquez sur **Add webhook**

### 2. Vérifier la configuration

Après chaque exécution de workflow, GitHub envoie un POST vers Kura. Les runs apparaissent dans l’interface sans synchronisation manuelle.

### 3. Récupérer l’URL du webhook dans l’interface

Dans la page **Pipelines** de Kura, ouvrez la section « Connecter un dépôt GitHub », puis « Option temps réel ». L’URL complète du webhook à copier y est affichée.

---

## Option alternative : Synchronisation par API (fallback)

Si les webhooks ne peuvent pas être utilisés (réseau, pare-feu, etc.), la synchronisation par API reste disponible.

### 1. Créer un Personal Access Token

1. GitHub : **Settings** → **Developer settings** → **Personal access tokens**
2. Générez un token avec les scopes `repo` ou au minimum `actions:read`
3. Copiez le token

### 2. Configurer dans Kura

1. Page **Pipelines** → section **Connecter un dépôt GitHub**
2. Collez le token dans le champ « Token GitHub »
3. Indiquez les dépôts au format `owner/repo` (séparés par des virgules)
4. Cliquez sur **Enregistrer**

### 3. Synchronisation

- Les runs sont récupérés périodiquement ou manuellement via le bouton « Synchroniser »
- Cette option peut être plus lente que les webhooks

---

## Résumé

| Méthode         | Avantages                          | Inconvénients              |
|----------------|-------------------------------------|----------------------------|
| **Webhooks**   | Temps réel, pas de polling          | Configuration sur GitHub   |
| **Sync API**   | Simple, sans config externe         | Délai, basé sur le polling |

**Recommandation** : utiliser les webhooks en priorité pour une expérience temps réel, et la sync API en complément ou en secours.
