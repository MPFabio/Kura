# Configuration AWX pour le service Ansible

Le service Ansible peut fonctionner **sans AWX** (mode dégradé), mais pour utiliser toutes les fonctionnalités, vous devez configurer une instance AWX ou Ansible Tower.

## Option 1 : Service sans AWX (par défaut)

Le service démarre sans AWX et fonctionne en mode dégradé :
- Les endpoints retourneront des erreurs 503 si AWX n'est pas configuré
- L'analyse de playbooks fonctionne toujours
- Les webhooks et métriques fonctionnent

```bash
# Démarrer uniquement le service Ansible
docker compose up -d ansible-service
```

## Option 2 : Utiliser une instance AWX existante

Si vous avez déjà une instance AWX/Tower en cours d'exécution :

1. Configurez les variables d'environnement dans `docker-compose.yml` :

```yaml
ansible-service:
  environment:
    ANSIBLE_TOWER_URL: http://votre-awx:8080  # URL de votre instance
    ANSIBLE_TOWER_USERNAME: admin
    ANSIBLE_TOWER_PASSWORD: votre-mot-de-passe
    ANSIBLE_TOWER_VERIFY_SSL: "false"
```

2. Redémarrez le service :

```bash
docker compose up -d ansible-service
```

## Option 3 : Installer AWX avec Docker Compose

AWX n'a plus d'image Docker officielle simple. Voici les options :

### Option 3a : Utiliser le docker-compose officiel d'AWX

1. Cloner le repo AWX :

```bash
git clone https://github.com/ansible/awx.git
cd awx/tools/docker-compose
```

2. Suivre les instructions pour construire et démarrer AWX

3. Configurer le service Ansible pour pointer vers cette instance

### Option 3b : Utiliser AWX Operator (Kubernetes)

Si vous utilisez Kubernetes/KinD :

```bash
# Installer AWX Operator
kubectl apply -f https://raw.githubusercontent.com/ansible/awx-operator/devel/deploy/awx-operator.yaml

# Créer une instance AWX
kubectl apply -f - <<EOF
apiVersion: awx.ansible.com/v1beta1
kind: AWX
metadata:
  name: awx
  namespace: kura
spec:
  service_type: LoadBalancer
EOF
```

### Option 3c : Utiliser une image alternative

Certaines images alternatives existent, mais elles ne sont pas officiellement supportées :

```yaml
# Exemple avec une image alternative (non testée)
awx:
  image: geerlingguy/awx:latest  # Image communautaire
  # ... reste de la configuration
```

## Vérification

Une fois AWX configuré, vérifiez la connexion :

```bash
# Vérifier que le service peut se connecter à AWX
curl http://localhost:8083/health

# La réponse devrait indiquer "ansible_tower_configured": true
```

## Test sans AWX

Même sans AWX, vous pouvez tester certaines fonctionnalités :

```bash
# Analyser un playbook (fonctionne sans AWX)
curl -X POST http://localhost:8083/api/v1/ansible/playbooks/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "playbook_content": "---\n- name: Test\n  hosts: localhost\n  tasks:\n    - debug:\n        msg: Hello"
  }'

# Vérifier les métriques Prometheus (fonctionne sans AWX)
curl http://localhost:8083/metrics
```

## Documentation AWX

- [AWX GitHub](https://github.com/ansible/awx)
- [AWX Docker Compose Setup](https://github.com/ansible/awx/tree/devel/tools/docker-compose)
- [AWX Operator](https://github.com/ansible/awx-operator)
