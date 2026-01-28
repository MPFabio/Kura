# Quick Start - Service Ansible avec AWX

Guide rapide pour démarrer et tester le service Ansible avec AWX dans un codespace.

## 🚀 Démarrage rapide (5 minutes)

### 1. Créer le réseau Docker (si pas déjà créé)

```bash
docker network create kura-network 2>/dev/null || true
```

### 2. Démarrer AWX et le service Ansible

```bash
# Depuis la racine du projet
docker-compose -f docker-compose.awx.yml up -d
```

**⏱️ Attendre 2-3 minutes** que AWX démarre complètement.

### 3. Vérifier que tout fonctionne

```bash
# Vérifier AWX
curl http://localhost:8080/api/v2/ping/

# Vérifier le service Ansible
curl http://localhost:8083/health

# Lancer les tests
cd services/ansible-service
chmod +x test-service.sh
./test-service.sh
```

### 4. Accéder aux interfaces

- **AWX Web UI** : http://localhost:8080 (admin/admin)
- **Service Ansible API** : http://localhost:8083/docs
- **Métriques Prometheus** : http://localhost:8083/metrics

## 📝 Test manuel rapide

### Créer une organisation dans AWX

```bash
curl -X POST http://localhost:8080/api/v2/organizations/ \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Org"}'
```

### Lister les organisations via le service

```bash
curl http://localhost:8083/api/v1/ansible/organizations
```

### Analyser un playbook

```bash
curl -X POST http://localhost:8083/api/v1/ansible/playbooks/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "playbook_content": "---\n- name: Test\n  hosts: localhost\n  tasks:\n    - debug:\n        msg: Hello"
  }'
```

## 🔧 Configuration dans un Codespace GitHub

### Variables d'environnement pour le service

Créez `services/ansible-service/.env` :

```bash
ANSIBLE_SERVICE_PORT=8083
ENV=development
LOG_LEVEL=info
REDIS_HOST=redis
REDIS_PORT=6379
ANSIBLE_TOWER_URL=http://awx:8080
ANSIBLE_TOWER_USERNAME=admin
ANSIBLE_TOWER_PASSWORD=admin
ANSIBLE_TOWER_VERIFY_SSL=false
```

### Exposer les ports dans le codespace

Dans l'interface GitHub Codespaces :
1. Onglet "Ports"
2. Exposer les ports :
   - **8080** (AWX) - Public
   - **8083** (Service Ansible) - Public
   - **6379** (Redis) - Private

### URLs publiques

Les URLs seront du type :
- `https://your-codespace-8080.preview.app.github.dev` (AWX)
- `https://your-codespace-8083.preview.app.github.dev` (Service Ansible)

## 🐛 Dépannage

### AWX ne démarre pas

```bash
# Vérifier les logs
docker-compose -f docker-compose.awx.yml logs -f awx

# Vérifier PostgreSQL
docker-compose -f docker-compose.awx.yml exec postgres-awx pg_isready
```

### Le service ne peut pas se connecter à AWX

1. Vérifier que AWX est prêt :
   ```bash
   curl http://localhost:8080/api/v2/ping/
   ```

2. Vérifier les logs du service :
   ```bash
   docker logs kura-ansible-service
   ```

3. Vérifier la configuration :
   ```bash
   docker exec kura-ansible-service env | grep ANSIBLE
   ```

### Redis non disponible

Le service fonctionne sans Redis mais sans cache. Pour démarrer Redis :

```bash
docker-compose up -d redis
```

## 📚 Prochaines étapes

1. **Créer des données de test dans AWX** :
   - Organisation
   - Inventaire avec hôtes
   - Projet Git
   - Template de job

2. **Tester les endpoints** :
   - Voir `TEST.md` pour les exemples complets

3. **Explorer la documentation** :
   - Swagger UI : http://localhost:8083/docs
   - ReDoc : http://localhost:8083/redoc

4. **Tester le streaming WebSocket** :
   ```bash
   # Installer wscat
   npm install -g wscat
   
   # Se connecter au stream d'un job
   wscat -c "ws://localhost:8083/api/v1/ansible/jobs/1/stream"
   ```

## 🛑 Arrêter les services

```bash
docker-compose -f docker-compose.awx.yml down

# Pour supprimer aussi les volumes (données AWX)
docker-compose -f docker-compose.awx.yml down -v
```
