# Configuration Grafana pour les métriques Ansible

Ce guide explique comment accéder et utiliser Grafana pour visualiser les métriques du service Ansible.

## Accès à Grafana

1. **URL** : http://localhost:3000
2. **Identifiants par défaut** :
   - Username: `admin`
   - Password: `admin`

## Configuration automatique

Le datasource Prometheus est automatiquement configuré lors du démarrage de Grafana grâce au provisioning.

## Accéder au dashboard Ansible

### Option 1 : Via le menu (recommandé)

1. Connectez-vous à Grafana (http://localhost:3000)
2. Dans le menu de gauche, cliquez sur **"Dashboards"** (icône carrée avec 4 petits carrés)
3. Cliquez sur **"Browse"** ou **"Parcourir"**
4. Vous devriez voir le dashboard **"Ansible Service - Métriques"**
5. Cliquez dessus pour l'ouvrir

### Option 2 : Via l'URL directe

Accédez directement à :
```
http://localhost:3000/d/ansible-service/ansible-service-metriques
```

### Option 3 : Recherche

1. Dans Grafana, utilisez la barre de recherche en haut (icône de loupe)
2. Tapez "Ansible"
3. Sélectionnez le dashboard dans les résultats

## Métriques disponibles

Le dashboard Ansible affiche :

1. **Requêtes API par seconde** : Graphique des requêtes API du service
2. **Statut des requêtes API** : Répartition par code de statut (200, 404, 500, etc.)
3. **Durée des requêtes API (p95)** : Temps de réponse au 95e percentile
4. **Jobs lancés** : Taux de jobs Ansible lancés par template
5. **Requêtes vers Ansible Tower** : Métriques des appels API vers Ansible Tower/AWX
6. **Webhooks reçus** : Taux de webhooks reçus par type d'événement
7. **Connexions WebSocket actives** : Nombre de connexions WebSocket actives
8. **Cache Hits/Misses** : Performance du cache Redis

## Vérifier que Prometheus collecte les métriques

1. Accédez à Prometheus : http://localhost:9090
2. Dans la barre de recherche, tapez : `ansible_api_requests_total`
3. Cliquez sur "Execute"
4. Si vous voyez des résultats, Prometheus collecte bien les métriques

## Dépannage

### Le dashboard n'apparaît pas

1. Vérifiez que les volumes sont bien montés dans `docker-compose.yml`
2. Redémarrez Grafana : `docker-compose restart grafana`
3. Vérifiez les logs : `docker-compose logs grafana`

### Aucune donnée dans le dashboard

1. Vérifiez que Prometheus collecte les métriques :
   - http://localhost:9090/targets
   - Le target `ansible-service:8083` doit être "UP"
2. Vérifiez que le service Ansible expose les métriques :
   - http://localhost:8083/metrics
   - Vous devriez voir des métriques commençant par `ansible_`
3. Vérifiez que Prometheus peut atteindre le service :
   - `docker-compose exec prometheus wget -qO- http://ansible-service:8083/metrics`

### Le datasource Prometheus n'est pas configuré

1. Allez dans **Configuration** → **Data sources**
2. Cliquez sur **"Add data source"**
3. Sélectionnez **"Prometheus"**
4. URL : `http://prometheus:9090`
5. Cliquez sur **"Save & Test"**

## Personnalisation

Vous pouvez modifier le dashboard en :

1. Ouvrant le dashboard dans Grafana
2. Cliquant sur l'icône d'engrenage en haut à droite
3. Sélectionnant **"Edit JSON"**
4. Modifiant le JSON selon vos besoins
5. Sauvegardant les modifications

Les modifications seront perdues si vous recréez le conteneur. Pour les rendre permanentes, modifiez le fichier `infrastructure/docker/grafana/dashboards/ansible-service.json`.
