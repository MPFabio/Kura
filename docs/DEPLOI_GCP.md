# Déployer Kura sur GCP via GitHub Actions

Ce guide explique comment déployer Kura (frontend + backend) sur une VM GCP via un pipeline GitHub Actions.

## Prérequis

- Compte GCP avec un projet
- `gcloud` et `terraform` installés en local
- Un dépôt GitHub pour ModulOps

## 1. Créer l’infrastructure avec Terraform

### 1.1 Backend GCS pour le tfstate (obligatoire en équipe)

Le state contient la clé SSH privée (`deploy_private_key`). Stocker le tfstate dans GCS (versionné) est fortement recommandé — obligatoire si plusieurs personnes appliquent Terraform.

```bash
cd infrastructure/terraform/gcp

# Créer le bucket GCS (une seule fois)
gsutil mb -p VOTRE_PROJET -l EU gs://VOTRE_PROJET-kura-tfstate
gsutil versioning set on gs://VOTRE_PROJET-kura-tfstate

# Activer le backend
cp backend.tf.example backend.tf
# Éditer backend.tf : remplacer VOTRE_PROJET-kura-tfstate par le nom réel du bucket
```

### 1.2 Appliquer Terraform

```bash
cp terraform.tfvars.example terraform.tfvars
# Éditer terraform.tfvars avec votre projet GCP

terraform init
terraform plan
terraform apply
```

Notez :
- **IP externe** : output `vm_external_ip`
- **Clé privée SSH** : `terraform output -raw deploy_private_key` → à copier dans le secret GitHub `GCP_SSH_PRIVATE_KEY`
- **Compte de service** : output `github_actions_sa_email`

### Créer une clé pour le compte de service GitHub Actions

1. Console GCP → **IAM & Admin** → **Service Accounts**
2. Ouvrir le compte `kura-github-actions`
3. **Keys** → **Add Key** → **Create new key** → JSON
4. Télécharger le fichier JSON

## 2. Configurer les secrets GitHub

Dans **Settings** → **Secrets and variables** → **Actions**, ajouter :

| Secret           | Description                                                |
|------------------|------------------------------------------------------------|
| `GCP_SA_KEY`     | Contenu du fichier JSON du compte de service               |
| `GCP_PROJECT_ID` | ID du projet GCP                                           |
| `GCP_VM_IP`      | IP externe de la VM (`terraform output vm_external_ip`)    |
| `GCP_SSH_PRIVATE_KEY` | Sortie de `terraform output -raw deploy_private_key` |

## 3. Variables de dépôt (optionnel)

| Variable             | Description                                                       |
|----------------------|-------------------------------------------------------------------|
| `VITE_API_BASE_URL`  | URL de l’API pour le frontend (ex: `https://VOTRE_IP:8000`)       |

Si non définie, le frontend utilisera `http://localhost:8000` (à remplacer après déploiement).

## 4. Déploiement

- **Automatique** : à chaque push sur `main`
- **Manuel** : **Actions** → **Deploy on GCP** → **Run workflow**

## 5. Accès après déploiement

- **Frontend** : `http://VOTRE_IP` (port 80)
- **API** : `http://VOTRE_IP:8000`

## 6. Premier déploiement

Après le premier `terraform apply`, attendre 2–3 minutes que le script de démarrage de la VM installe Docker et gcloud. Ensuite, lancer un déploiement manuel ou pousser sur `main`.

## 7. Ressources créées

- **VM** : e2-micro (1 Go RAM, gratuit dans les régions éligibles)
- **Artifact Registry** : dépôt `kura` pour les images Docker
- **Pare-feu** : SSH (22), HTTP (80), API (8000)

## Dépannage

### VM en manque de mémoire (e2-micro)

Passer à `e2-small` dans `terraform.tfvars` :

```hcl
machine_type = "e2-small"
```

### Images non trouvées sur la VM

Vérifier que la VM a bien le rôle `roles/artifactregistry.reader` (géré par Terraform).

### Erreur SSH

Vérifier que `GCP_SSH_PRIVATE_KEY` contient bien la sortie de `terraform output -raw deploy_private_key`. La clé est générée par Terraform et injectée sur la VM.
