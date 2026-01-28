# Quick Start - Service Ansible avec AWX

Guide rapide pour démarrer et tester le service Ansible avec AWX via Docker Compose dans un codespace.

## 🚀 Démarrage rapide (5 minutes)

### 1. Démarrer les services

```bash
# Depuis la racine du projet
docker-compose up -d

# Ou démarrer uniquement AWX et le service Ansible
docker-compose up -d awx-postgres awx-memcached awx ansible-service
```

**⏱️ AWX peut prendre 3-5 minutes** à démarrer complètement.

### 2. Vérifier que tout fonctionne

```bash
# Vérifier AWX (attendre qu'il soit prêt)
# Note: AWX est sur le port 8084 car 8080 est utilisé par auth-service
curl http://localhost:8084/api/v2/ping/

# Vérifier le service Ansible
curl http://localhost:8083/health

# Lancer les tests
cd services/ansible-service
bash test-service.sh
```

### 6. Accéder aux interfaces

- **AWX Web UI** : http://localhost:8084 (admin/admin) - *Note: port 8084 car 8080 est utilisé par auth-service*
- **Service Ansible API** : http://localhost:8083/docs
- **Métriques Prometheus** : http://localhost:8083/metrics

**Note** : Dans un codespace GitHub, vous pouvez exposer les ports via l'interface "Ports" pour obtenir des URLs publiques.

## 📝 Test manuel rapide

### Créer une organisation dans AWX

```bash
# Via port-forward
curl -X POST http://localhost:8084/api/v2/organizations/ \
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

### Exposer les ports dans le codespace

Dans l'interface GitHub Codespaces :
1. Onglet "Ports"
2. Exposer les ports :
   - **8080** (AWX) - Public
   - **8083** (Service Ansible) - Public

### URLs publiques

Les URLs seront du type :
- `https://your-codespace-8080.preview.app.github.dev` (AWX)
- `https://your-codespace-8083.preview.app.github.dev` (Service Ansible)

## 🐛 Dépannage

### AWX ne démarre pas

```bash
# Vérifier les logs
docker-compose logs -f awx

# Vérifier l'état des conteneurs
docker-compose ps

# Vérifier PostgreSQL
docker-compose exec awx-postgres pg_isready -U awx

# Redémarrer AWX
docker-compose restart awx
```

### Le service ne peut pas se connecter à AWX

1. Vérifier que AWX est prêt :
   ```bash
   curl http://localhost:8084/api/v2/ping/
   ```

2. Vérifier les logs du service :
   ```bash
   docker-compose logs -f ansible-service
   ```

3. Vérifier la configuration :
   ```bash
   docker-compose exec ansible-service env | grep ANSIBLE
   ```

### Rebuilder le service après modifications

```bash
# Rebuilder l'image
docker-compose build ansible-service

# Redémarrer le service
docker-compose up -d ansible-service
```

### Redis non disponible

Le service fonctionne sans Redis mais sans cache. Vérifier Redis :

```bash
docker-compose ps redis
docker-compose logs redis
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
# Arrêter tous les services
docker-compose down

# Ou arrêter uniquement AWX et le service Ansible
docker-compose stop awx ansible-service awx-postgres awx-memcached

# Pour supprimer aussi les volumes (données AWX)
docker-compose down -v
```
