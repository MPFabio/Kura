# kubectl avec un cluster GKE

Si vous voyez :

```text
exec: executable gke-gcloud-auth-plugin not found
```

c’est que votre `kubeconfig` pointe vers un cluster **Google Kubernetes Engine (GKE)**. GKE utilise un plugin d’authentification qui doit être installé séparément.

## 1. Installer le plugin GKE (obligatoire pour GKE)

### Avec gcloud CLI (recommandé)

**Windows (PowerShell ou CMD)**  
À condition d’avoir [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) installé :

```powershell
gcloud components install gke-gcloud-auth-plugin
```

**macOS (Homebrew)**  
Si vous utilisez le `gcloud` fourni par Homebrew, le plugin est souvent inclus. Sinon :

```bash
gcloud components install gke-gcloud-auth-plugin
```

**Linux**  
Plugin standalone (sans gcloud) :

```bash
# Par exemple sur Debian/Ubuntu
curl -LO "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-cli-linux-x86_64.tar.gz"
# Puis utiliser gcloud components install gke-gcloud-auth-plugin
# Ou installer le binaire seul : https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin
```

Après installation, vérifier que le binaire est dans le `PATH` :

```bash
gke-gcloud-auth-plugin --version
```

Puis rafraîchir les identifiants du cluster :

```bash
gcloud container clusters get-credentials NOM_DU_CLUSTER --region REGION --project ID_PROJECT
```

Ensuite `kubectl apply -f k8s/ -R` (ou `kubectl apply -k infrastructure/k8s/`) pourra s’exécuter normalement.

## 2. Ne pas utiliser GKE (cluster local / autre cloud)

Si vous ne ciblez **pas** un cluster GKE mais que votre `kubeconfig` pointe encore dessus :

- Utiliser le contexte d’un autre cluster :
  ```bash
  kubectl config get-contexts
  kubectl config use-context VOTRE_CONTEXTE_LOCAL
  ```
- Ou désigner un autre fichier kubeconfig pour un cluster local :
  ```bash
  set KUBECONFIG=C:\chemin\vers\votre\kubeconfig
  kubectl apply -k infrastructure/k8s/
  ```

## 3. Appliquer les manifests du projet

Une fois l’accès au cluster correct (GKE ou autre) :

```bash
# À la racine du dépôt
kubectl apply -k infrastructure/k8s/
```

Ou, si vous êtes dans `infrastructure` :

```bash
kubectl apply -f k8s/ -R
```

Référence : [Cluster access for kubectl (GKE)](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin).
