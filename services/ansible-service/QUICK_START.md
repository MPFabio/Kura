# Quick Start - Service Ansible avec AWX dans KinD

Guide rapide pour démarrer et tester le service Ansible avec AWX déployé sur KinD dans un codespace.

## 🚀 Démarrage rapide (10 minutes)

### 1. Vérifier le cluster KinD

```bash
# Vérifier que le cluster existe
kind get clusters

# Si le cluster n'existe pas, le créer
./scripts/setup-k8s.sh
```

### 2. Builder et déployer le service Ansible

```bash
# Builder l'image du service Ansible
cd services/ansible-service
docker build -t ansible-service:latest .

# Charger l'image dans KinD
kind load docker-image ansible-service:latest --name kura

# Retourner à la racine
cd ../..
```

### 3. Déployer AWX et le service Ansible

```bash
# Déployer tous les services (y compris AWX et ansible-service)
kubectl apply -k infrastructure/k8s/

# Attendre que les services soient prêts
kubectl wait --for=condition=ready --timeout=600s pod -l app=awx -n kura
kubectl wait --for=condition=ready --timeout=300s pod -l app=ansible-service -n kura
```

**⏱️ AWX peut prendre 3-5 minutes** à démarrer complètement.

### 4. Configurer le port-forward

```bash
# Dans des terminaux séparés ou en arrière-plan

# Terminal 1: AWX
kubectl port-forward svc/awx 8080:8080 -n kura

# Terminal 2: Service Ansible
kubectl port-forward svc/ansible-service 8083:8083 -n kura
```

### 5. Vérifier que tout fonctionne

```bash
# Vérifier AWX
curl http://localhost:8080/api/v2/ping/

# Vérifier le service Ansible
curl http://localhost:8083/health

# Lancer les tests
cd services/ansible-service
bash test-service.sh
```

### 6. Accéder aux interfaces

Après avoir configuré le port-forward :

- **AWX Web UI** : http://localhost:8080 (admin/admin)
- **Service Ansible API** : http://localhost:8083/docs
- **Métriques Prometheus** : http://localhost:8083/metrics

**Note** : Dans un codespace GitHub, vous pouvez exposer les ports via l'interface "Ports" pour obtenir des URLs publiques.

## 📝 Test manuel rapide

### Créer une organisation dans AWX

```bash
# Via port-forward
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

## 🔧 Configuration dans un Codespace GitHub avec KinD

### Exposer les ports dans le codespace

Dans l'interface GitHub Codespaces :
1. Onglet "Ports"
2. Exposer les ports après avoir fait le port-forward :
   - **8080** (AWX) - Public
   - **8083** (Service Ansible) - Public

### URLs publiques

Les URLs seront du type :
- `https://your-codespace-8080.preview.app.github.dev` (AWX)
- `https://your-codespace-8083.preview.app.github.dev` (Service Ansible)

### Script de port-forward automatique

Créez `scripts/port-forward-services.sh` :

```bash
#!/bin/bash
# Port-forward pour AWX et le service Ansible

kubectl port-forward svc/awx 8080:8080 -n kura &
kubectl port-forward svc/ansible-service 8083:8083 -n kura &

echo "Port-forward actif :"
echo "  - AWX: http://localhost:8080"
echo "  - Service Ansible: http://localhost:8083"
echo ""
echo "Appuyez sur Ctrl+C pour arrêter"
wait
```

## 🐛 Dépannage

### AWX ne démarre pas

```bash
# Vérifier les logs
kubectl logs -f deployment/awx -n kura

# Vérifier l'état des pods
kubectl get pods -n kura | grep awx

# Vérifier PostgreSQL
kubectl exec -it deployment/awx-postgres -n kura -- pg_isready -U awx

# Vérifier les événements
kubectl describe pod -l app=awx -n kura
```

### Le service ne peut pas se connecter à AWX

1. Vérifier que AWX est prêt depuis le cluster :
   ```bash
   kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- curl http://awx:8080/api/v2/ping/
   ```

2. Vérifier les logs du service :
   ```bash
   kubectl logs -f deployment/ansible-service -n kura
   ```

3. Vérifier la configuration :
   ```bash
   kubectl get configmap ansible-service-config -n kura -o yaml
   ```

### Rebuilder le service après modifications

```bash
cd services/ansible-service
docker build -t ansible-service:latest .
kind load docker-image ansible-service:latest --name kura
kubectl rollout restart deployment/ansible-service -n kura
```

### Redis non disponible

Le service fonctionne sans Redis mais sans cache. Vérifier Redis :

```bash
kubectl get pods -n kura | grep redis
kubectl logs deployment/redis -n kura
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
# Arrêter les port-forwards (Ctrl+C dans les terminaux)

# Supprimer les déploiements
kubectl delete -k infrastructure/k8s/

# Ou supprimer uniquement AWX et le service Ansible
kubectl delete deployment awx awx-postgres awx-memcached ansible-service -n kura
kubectl delete svc awx awx-postgres awx-memcached ansible-service -n kura
kubectl delete pvc awx-pvc awx-postgres-pvc -n kura
```
