# Guide de test du service Ansible avec AWX

Ce guide explique comment tester le service Ansible dans un codespace avec AWX.

## Prérequis

- Codespace GitHub avec Docker
- Accès à Internet pour télécharger les images Docker

## Étape 1 : Installer AWX dans le codespace

### Option A : AWX via Docker Compose (recommandé pour le développement)

Créez un fichier `docker-compose.awx.yml` à la racine du projet :

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:13
    environment:
      POSTGRES_DB: awx
      POSTGRES_USER: awx
      POSTGRES_PASSWORD: awxpassword
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  memcached:
    image: memcached:alpine
    ports:
      - "11211:11211"

  awx:
    image: ansible/awx:latest
    depends_on:
      - postgres
      - memcached
    environment:
      AWX_ADMIN_USER: admin
      AWX_ADMIN_PASSWORD: admin
      DATABASE_HOST: postgres
      DATABASE_PORT: 5432
      DATABASE_NAME: awx
      DATABASE_USER: awx
      DATABASE_PASSWORD: awxpassword
      MEMCACHED_HOST: memcached
      MEMCACHED_PORT: 11211
    ports:
      - "8080:8080"
    volumes:
      - awx_data:/var/lib/awx

volumes:
  postgres_data:
  awx_data:
```

Lancez AWX :

```bash
docker-compose -f docker-compose.awx.yml up -d
```

Attendez que AWX soit prêt (peut prendre 2-3 minutes) :

```bash
# Vérifier les logs
docker-compose -f docker-compose.awx.yml logs -f awx

# Vérifier que le service répond
curl http://localhost:8080/api/v2/ping/
```

### Option B : AWX via Kubernetes (si vous avez kubectl)

```bash
# Installer AWX Operator
kubectl apply -f https://raw.githubusercontent.com/ansible/awx-operator/devel/deploy/awx-operator.yaml

# Créer une instance AWX
kubectl apply -f - <<EOF
apiVersion: awx.ansible.com/v1beta1
kind: AWX
metadata:
  name: awx
spec:
  service_type: NodePort
EOF
```

## Étape 2 : Configurer le service Ansible

### Créer un fichier `.env` dans `services/ansible-service/`

```bash
cd services/ansible-service

cat > .env <<EOF
# Serveur
ANSIBLE_SERVICE_PORT=8083
ENV=development
LOG_LEVEL=info

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
ANSIBLE_CACHE_TTL=300

# Ansible Tower/AWX
ANSIBLE_TOWER_URL=http://localhost:8080
ANSIBLE_TOWER_USERNAME=admin
ANSIBLE_TOWER_PASSWORD=admin
ANSIBLE_TOWER_VERIFY_SSL=false
EOF
```

### Installer les dépendances Python

```bash
cd services/ansible-service
pip install -r requirements.txt
```

## Étape 3 : Démarrer Redis (si pas déjà démarré)

```bash
# Via Docker
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Ou vérifier si Redis est déjà disponible
redis-cli ping
```

## Étape 4 : Démarrer le service Ansible

```bash
cd services/ansible-service

# Option 1 : Directement avec Python
python main.py

# Option 2 : Avec uvicorn (recommandé pour le développement)
uvicorn main:app --host 0.0.0.0 --port 8083 --reload
```

Le service devrait démarrer sur `http://localhost:8083`

## Étape 5 : Tester les endpoints

### Vérifier la santé du service

```bash
curl http://localhost:8083/health
```

Réponse attendue :
```json
{
  "status": "ok",
  "service": "ansible-service",
  "ansible_tower_configured": true,
  "redis_available": true
}
```

### Tester l'API Ansible

#### 1. Lister les jobs

```bash
curl http://localhost:8083/api/v1/ansible/jobs
```

#### 2. Lister les inventaires

```bash
curl http://localhost:8083/api/v1/ansible/inventories
```

#### 3. Lister les templates de jobs

```bash
curl http://localhost:8083/api/v1/ansible/job-templates
```

#### 4. Lister les organisations

```bash
curl http://localhost:8083/api/v1/ansible/organizations
```

#### 5. Lister les credentials

```bash
curl http://localhost:8083/api/v1/ansible/credentials
```

#### 6. Analyser un playbook

```bash
curl -X POST http://localhost:8083/api/v1/ansible/playbooks/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "playbook_content": "---
- name: Test playbook
  hosts: localhost
  tasks:
    - name: Test task
      debug:
        msg: Hello World"
  }'
```

#### 7. Lancer un template de job

```bash
# D'abord, récupérer l'ID d'un template
TEMPLATE_ID=$(curl -s http://localhost:8083/api/v1/ansible/job-templates | jq -r '.results[0].id')

# Lancer le template
curl -X POST "http://localhost:8083/api/v1/ansible/job-templates/$TEMPLATE_ID/launch" \
  -H "Content-Type: application/json" \
  -d '{
    "extra_vars": {
      "variable1": "value1"
    }
  }'
```

#### 8. Streamer les logs d'un job (WebSocket)

```bash
# Utiliser wscat ou un client WebSocket
# Installer wscat : npm install -g wscat

JOB_ID=1
wscat -c "ws://localhost:8083/api/v1/ansible/jobs/$JOB_ID/stream"
```

#### 9. Vérifier les métriques Prometheus

```bash
curl http://localhost:8083/metrics
```

### Tester les webhooks

```bash
# Simuler un webhook d'Ansible Tower
curl -X POST http://localhost:8083/api/v1/webhooks/ansible-tower \
  -H "Content-Type: application/json" \
  -d '{
    "event": "job_started",
    "job_id": 123,
    "status": "running",
    "job_template_id": 42
  }'
```

## Étape 6 : Accéder à la documentation interactive

Ouvrez dans votre navigateur :

- **Swagger UI** : http://localhost:8083/docs
- **ReDoc** : http://localhost:8083/redoc

## Étape 7 : Créer des données de test dans AWX

### Via l'interface web AWX

1. Accédez à http://localhost:8080
2. Connectez-vous avec `admin` / `admin`
3. Créez :
   - Une organisation
   - Un inventaire avec quelques hôtes
   - Un projet (GitHub, GitLab, ou manuel)
   - Un template de job
   - Un credential

### Via l'API AWX directement

```bash
# Créer une organisation
curl -X POST http://localhost:8080/api/v2/organizations/ \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Organization",
    "description": "Organization de test"
  }'

# Créer un inventaire
curl -X POST http://localhost:8080/api/v2/inventories/ \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Inventory",
    "organization": 1
  }'
```

## Dépannage

### AWX ne démarre pas

```bash
# Vérifier les logs
docker-compose -f docker-compose.awx.yml logs awx

# Vérifier que PostgreSQL est prêt
docker-compose -f docker-compose.awx.yml exec postgres pg_isready
```

### Le service Ansible ne peut pas se connecter à AWX

1. Vérifier que AWX est accessible :
   ```bash
   curl http://localhost:8080/api/v2/ping/
   ```

2. Vérifier les credentials dans `.env`

3. Vérifier les logs du service :
   ```bash
   # Les logs devraient montrer les erreurs de connexion
   ```

### Redis n'est pas disponible

Le service fonctionnera sans Redis mais sans cache. Pour démarrer Redis :

```bash
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

## Script de test complet

Créez un fichier `test-service.sh` :

```bash
#!/bin/bash

set -e

echo "🧪 Test du service Ansible"

# Vérifier la santé
echo "1. Vérification de la santé..."
curl -s http://localhost:8083/health | jq

# Lister les jobs
echo "2. Liste des jobs..."
curl -s http://localhost:8083/api/v1/ansible/jobs | jq '.count'

# Lister les inventaires
echo "3. Liste des inventaires..."
curl -s http://localhost:8083/api/v1/ansible/inventories | jq '.count'

# Lister les organisations
echo "4. Liste des organisations..."
curl -s http://localhost:8083/api/v1/ansible/organizations | jq '.count'

# Métriques Prometheus
echo "5. Métriques Prometheus..."
curl -s http://localhost:8083/metrics | grep ansible | head -5

echo "✅ Tests terminés"
```

Rendez-le exécutable et lancez-le :

```bash
chmod +x test-service.sh
./test-service.sh
```

## Notes importantes

1. **Ports utilisés** :
   - AWX : 8080
   - Service Ansible : 8083
   - Redis : 6379
   - PostgreSQL : 5432

2. **Dans un codespace GitHub** :
   - Les ports doivent être exposés via les "Ports" dans l'interface
   - Utilisez les URLs publiques générées par GitHub Codespaces

3. **Performance** :
   - AWX peut prendre plusieurs minutes à démarrer la première fois
   - Le service Ansible démarre en quelques secondes

4. **Sécurité** :
   - En production, changez les mots de passe par défaut
   - Utilisez HTTPS
   - Configurez l'authentification JWT
