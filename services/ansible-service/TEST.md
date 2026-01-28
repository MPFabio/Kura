# Guide de test du service Ansible avec AWX dans KinD

Ce guide explique comment tester le service Ansible dans un codespace avec AWX déployé sur KinD (Kubernetes in Docker).

## Prérequis

- Codespace GitHub avec Docker
- KinD installé et cluster créé
- kubectl configuré pour le cluster KinD
- Accès à Internet pour télécharger les images Docker

## Étape 1 : Vérifier le cluster KinD

### Vérifier que le cluster KinD est actif

```bash
# Vérifier le cluster
kubectl cluster-info --context kind-kura

# Vérifier les nodes
kubectl get nodes
```

### Déployer AWX et le service Ansible sur KinD

Les manifests Kubernetes sont déjà configurés dans `infrastructure/k8s/`. 

```bash
# Builder l'image du service Ansible et la charger dans KinD
cd services/ansible-service
docker build -t ansible-service:latest .
kind load docker-image ansible-service:latest --name kura

# Déployer tous les services (y compris AWX et ansible-service)
cd ../..
kubectl apply -k infrastructure/k8s/

# Attendre que les pods soient prêts
kubectl wait --for=condition=ready --timeout=300s pod -l app=awx-postgres -n kura
kubectl wait --for=condition=ready --timeout=300s pod -l app=awx-memcached -n kura
kubectl wait --for=condition=ready --timeout=600s pod -l app=awx -n kura
kubectl wait --for=condition=ready --timeout=300s pod -l app=ansible-service -n kura
```

**⏱️ AWX peut prendre 3-5 minutes** à démarrer complètement. Surveillez les logs :

```bash
# Vérifier les logs d'AWX
kubectl logs -f deployment/awx -n kura

# Vérifier que AWX répond
kubectl port-forward svc/awx 8080:8080 -n kura &
curl http://localhost:8080/api/v2/ping/
```

## Étape 2 : Accéder aux services

### Port-forward pour accéder aux services depuis l'extérieur du cluster

```bash
# Dans des terminaux séparés ou en arrière-plan

# AWX (port 8080)
kubectl port-forward svc/awx 8080:8080 -n kura

# Service Ansible (port 8083)
kubectl port-forward svc/ansible-service 8083:8083 -n kura
```

### Ou utiliser les services LoadBalancer (dans KinD)

Si vous avez configuré un LoadBalancer pour KinD (comme MetalLB), les services seront accessibles directement via leurs IPs externes :

```bash
# Obtenir l'IP externe d'AWX
kubectl get svc awx -n kura

# Obtenir l'IP externe du service Ansible
kubectl get svc ansible-service -n kura
```

## Étape 3 : Tester les endpoints

### Vérifier la santé du service

```bash
# Si vous utilisez port-forward
curl http://localhost:8083/health

# Ou directement via kubectl
kubectl exec -it deployment/ansible-service -n kura -- curl http://localhost:8083/health
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

## Étape 4 : Accéder à la documentation interactive

Ouvrez dans votre navigateur (après avoir fait le port-forward) :

- **AWX Web UI** : http://localhost:8080 (admin/admin)
- **Service Ansible Swagger UI** : http://localhost:8083/docs
- **Service Ansible ReDoc** : http://localhost:8083/redoc

## Étape 5 : Créer des données de test dans AWX

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
kubectl logs -f deployment/awx -n kura

# Vérifier les pods
kubectl get pods -n kura | grep awx

# Vérifier que PostgreSQL est prêt
kubectl exec -it deployment/awx-postgres -n kura -- pg_isready -U awx

# Vérifier les événements
kubectl describe pod -l app=awx -n kura
```

### Le service Ansible ne peut pas se connecter à AWX

1. Vérifier que AWX est accessible depuis le cluster :
   ```bash
   # Depuis un pod dans le cluster
   kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- curl http://awx:8080/api/v2/ping/
   ```

2. Vérifier la configuration du service :
   ```bash
   kubectl get configmap ansible-service-config -n kura -o yaml
   ```

3. Vérifier les logs du service :
   ```bash
   kubectl logs -f deployment/ansible-service -n kura
   ```

### Redis n'est pas disponible

Le service fonctionnera sans Redis mais sans cache. Vérifier Redis :

```bash
kubectl get pods -n kura | grep redis
kubectl logs deployment/redis -n kura
```

### Rebuilder et redéployer le service Ansible

```bash
# Builder l'image
cd services/ansible-service
docker build -t ansible-service:latest .

# Charger dans KinD
kind load docker-image ansible-service:latest --name kura

# Redéployer
kubectl rollout restart deployment/ansible-service -n kura
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
