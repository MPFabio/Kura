# Dépannage du service Ansible

## Problème : "Connection reset by peer"

### Symptômes
- Le conteneur démarre mais `curl http://localhost:8083/health` échoue avec "Connection reset by peer"
- Le service semble démarrer puis crasher immédiatement

### Solutions

#### 1. Vérifier les logs du conteneur

```bash
docker logs kura-ansible-service --tail=100
```

Cela vous montrera l'erreur exacte qui fait crasher le service.

#### 2. Vérifier que Redis est accessible

```bash
# Vérifier que Redis est démarré
docker-compose ps redis

# Tester la connexion depuis le conteneur
docker-compose exec ansible-service ping -c 1 redis
```

#### 3. Rebuilder l'image après modifications

```bash
# Rebuilder l'image
docker-compose build ansible-service

# Redémarrer le service
docker-compose up -d ansible-service
```

#### 4. Tester le service localement (sans Docker)

```bash
cd services/ansible-service

# Installer les dépendances
pip install -r requirements.txt

# Démarrer le service
python main.py
```

Cela vous permettra de voir les erreurs directement dans le terminal.

#### 5. Vérifier les dépendances Python

```bash
# Dans le conteneur
docker-compose exec ansible-service python -c "import fastapi; import redis; import httpx; print('OK')"
```

#### 6. Vérifier la configuration

```bash
# Vérifier les variables d'environnement
docker-compose exec ansible-service env | grep ANSIBLE
```

## Problème : Le service démarre mais les endpoints retournent 503

### Cause
AWX n'est pas configuré ou n'est pas accessible.

### Solution
Le service fonctionne en mode dégradé sans AWX. Certaines fonctionnalités fonctionnent toujours :

```bash
# Analyser un playbook (fonctionne sans AWX)
curl -X POST http://localhost:8083/api/v1/ansible/playbooks/analyze \
  -H "Content-Type: application/json" \
  -d '{"playbook_content": "---\n- name: Test\n  hosts: localhost"}'

# Métriques Prometheus (fonctionne sans AWX)
curl http://localhost:8083/metrics
```

Pour utiliser toutes les fonctionnalités, configurez AWX (voir `AWX_SETUP.md`).

## Problème : Erreur d'import Python

### Symptômes
```
ModuleNotFoundError: No module named 'internal'
```

### Solution

Le problème vient du fait que Python ne trouve pas le module `internal`. Vérifiez que vous êtes dans le bon répertoire :

```bash
# Le service doit être démarré depuis services/ansible-service
cd services/ansible-service
python main.py
```

Ou utilisez Docker Compose qui gère cela automatiquement.

## Problème : Redis connection error

### Symptômes
```
redis.exceptions.ConnectionError: Error connecting to Redis
```

### Solution

```bash
# Démarrer Redis
docker-compose up -d redis

# Vérifier que Redis répond
docker-compose exec redis redis-cli ping
```

Le service fonctionnera sans Redis mais sans cache.

## Commandes utiles

```bash
# Voir les logs en temps réel
docker-compose logs -f ansible-service

# Redémarrer le service
docker-compose restart ansible-service

# Rebuilder et redémarrer
docker-compose up -d --build ansible-service

# Entrer dans le conteneur pour déboguer
docker-compose exec ansible-service bash

# Tester la connexion HTTP depuis le conteneur
docker-compose exec ansible-service curl http://localhost:8083/health
```
