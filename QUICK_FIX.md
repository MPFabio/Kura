# Guide de dépannage rapide

## Problème 1 : Le frontend ne montre pas les changements

### Solution

Le frontend doit être reconstruit pour prendre en compte les nouveaux fichiers.

```bash
# Arrêter le frontend
docker-compose stop frontend

# Reconstruire l'image du frontend
docker-compose build frontend

# Redémarrer le frontend
docker-compose up -d frontend
```

Ou en une seule commande :
```bash
docker-compose up -d --build frontend
```

### Vérification

1. Accédez à http://localhost:5173
2. Allez dans la section "Ansible" (menu de gauche)
3. Vous devriez voir les onglets : Jobs, Historique, Inventaires, Templates

### Si ça ne fonctionne toujours pas

Vérifiez les logs du frontend :
```bash
docker-compose logs frontend
```

Vérifiez que le service Ansible répond :
```bash
curl http://localhost:8083/health
```

Vérifiez que Kong route correctement :
```bash
curl http://localhost:8000/api/v1/ansible/jobs
```

## Problème 2 : Accéder aux métriques dans Grafana

### Étapes

1. **Accédez à Grafana** : http://localhost:3000
   - Username: `admin`
   - Password: `admin`

2. **Redémarrez Grafana** pour charger la nouvelle configuration :
   ```bash
   docker-compose restart grafana
   ```

3. **Accédez au dashboard Ansible** :
   - Dans le menu de gauche, cliquez sur **"Dashboards"** (icône avec 4 carrés)
   - Cliquez sur **"Browse"**
   - Cherchez **"Ansible Service - Métriques"**
   - Cliquez dessus

   Ou directement via l'URL :
   ```
   http://localhost:3000/d/ansible-service/ansible-service-metriques
   ```

### Vérifier que Prometheus collecte les métriques

1. Accédez à Prometheus : http://localhost:9090
2. Allez dans **Status** → **Targets**
3. Vérifiez que `ansible-service:8083` est **UP** (vert)
4. Si c'est **DOWN** (rouge), redémarrez Prometheus :
   ```bash
   docker-compose restart prometheus
   ```

### Tester les métriques directement

```bash
# Vérifier que le service expose les métriques
curl http://localhost:8083/metrics | grep ansible

# Vérifier que Prometheus peut les récupérer
curl http://localhost:9090/api/v1/query?query=ansible_api_requests_total
```

## Redémarrer tous les services modifiés

Pour appliquer tous les changements :

```bash
# Redémarrer Prometheus (nouvelle config)
docker-compose restart prometheus

# Redémarrer Grafana (nouveaux dashboards)
docker-compose restart grafana

# Reconstruire et redémarrer le frontend
docker-compose up -d --build frontend

# Redémarrer Kong (nouvelle route)
docker-compose restart kong
```

## Vérification complète

```bash
# Vérifier que tous les services sont UP
docker-compose ps

# Vérifier les logs du service Ansible
docker-compose logs ansible-service --tail=50

# Vérifier les métriques
curl http://localhost:8083/metrics | head -20

# Tester l'API via Kong
curl http://localhost:8000/api/v1/ansible/jobs
```
