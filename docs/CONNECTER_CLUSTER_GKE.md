# Connecter un cluster GKE à Kura

Comme dans Lens : tu récupères le kubeconfig, tu le colles dans Kura. En Docker, il faut en plus fournir une **clé de compte de service GCP** pour que le plugin d’auth fonctionne dans le conteneur.

---

## 1. Récupérer le kubeconfig

Avec `gcloud` (Cloud Shell ou machine où gcloud est installé) :

```bash
gcloud container clusters get-credentials projet-ynov-gke --region europe-west1 --project kura-devops
```

Copie le contenu **complet** (avec certificats) :
- **PowerShell** : `Get-Content $env:USERPROFILE\.kube\config -Raw` ou `kubectl config view --raw`
- **Bash** : `cat ~/.kube/config` ou `kubectl config view --raw`

Ne pas utiliser `kubectl config view` seul (sans `--raw`) : les certificats seraient masqués (DATA+OMITTED) et Kura ne pourrait pas se connecter.

---

## 2. Dans Kura (sans Docker)

**Kubernetes** → **Ajouter un cluster** → coller le kubeconfig → **Créer**.  
Le plugin GKE est inclus dans le k8s-service.

---

## 3. En Docker : clé GCP obligatoire pour GKE

En Docker, le conteneur n’a pas tes identifiants gcloud. Il faut un **fichier de clé (JSON) d’un compte de service GCP** avec accès au cluster GKE.

### Étapes

1. **Créer un compte de service (ou en réutiliser un)**  
   Console GCP → **IAM et administration** → **Comptes de service** → **Créer un compte de service**.  
   Rôle conseillé : **Utilisateur Kubernetes Engine** (ou **Kubernetes Engine Developer**).

2. **Créer une clé JSON**  
   Sur le compte de service → **Clés** → **Ajouter une clé** → **Créer une clé** → **JSON** → télécharger.

3. **Placer le fichier dans le projet**  
   Copier le JSON dans `secrets/gcp-sa.json` à la racine du projet Kura (créer le dossier `secrets` si besoin).  
   Ne pas commiter ce fichier (il doit être dans `.gitignore`).

4. **Démarrer / redémarrer le k8s-service**  
   Le `docker-compose` monte déjà `./secrets/gcp-sa.json` dans le conteneur et définit `GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp-sa.json`.  
   Redémarrer au moins le service K8s :
   ```bash
   docker compose up -d --build k8s-service
   ```

5. **Dans Kura**  
   **Kubernetes** → **Clusters** → ajouter le cluster en collant le kubeconfig (étape 1) → activer le cluster.

### Si tu ne veux pas utiliser de clé dans Docker

Tu peux générer un kubeconfig avec **token** (token court durée) : voir [Cluster access for kubectl (GKE)](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl) → « Générer un kubeconfig avec token ». Ce kubeconfig pourra être utilisé sans `GOOGLE_APPLICATION_CREDENTIALS`, mais le token devra être renouvelé régulièrement.
