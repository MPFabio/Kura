# Guide de Versionnement - Kura

Ce document explique comment gérer les versions du projet Kura.

## Format de Version

Kura utilise le [Semantic Versioning](https://semver.org/lang/fr/) (SemVer) :

- **MAJOR** (X.0.0) : Changements incompatibles avec les versions précédentes
- **MINOR** (0.X.0) : Nouvelles fonctionnalités rétrocompatibles
- **PATCH** (0.0.X) : Corrections de bugs rétrocompatibles

Exemple : `1.2.3` = Major 1, Minor 2, Patch 3

## Fichiers de Version

- **`VERSION`** : Fichier principal contenant la version actuelle
- **`frontend/package.json`** : Version du package npm du frontend
- **`CHANGELOG.md`** : Historique des changements par version

## Créer une Nouvelle Version

### Méthode 1 : Script Automatique (Recommandé)

#### Sur Windows (PowerShell) :
```powershell
# Incrémenter automatiquement
.\scripts\version.ps1 -Type patch    # 1.0.0 → 1.0.1
.\scripts\version.ps1 -Type minor    # 1.0.0 → 1.1.0
.\scripts\version.ps1 -Type major    # 1.0.0 → 2.0.0

# Version spécifique
.\scripts\version.ps1 -Version 1.2.3
```

#### Sur Linux/Mac (Bash) :
```bash
# Incrémenter automatiquement
./scripts/version.sh patch    # 1.0.0 → 1.0.1
./scripts/version.sh minor    # 1.0.0 → 1.1.0
./scripts/version.sh major    # 1.0.0 → 2.0.0

# Version spécifique
./scripts/version.sh 1.2.3
```

### Méthode 2 : Manuel

1. **Mettre à jour le fichier VERSION** :
   ```bash
   echo "1.0.0" > VERSION
   ```

2. **Mettre à jour frontend/package.json** :
   ```json
   {
     "version": "1.0.0"
   }
   ```

3. **Mettre à jour CHANGELOG.md** :
   Ajouter une nouvelle section pour la version :
   ```markdown
   ## [1.0.0] - 2026-01-20
   
   ### Ajouté
   - Nouvelle fonctionnalité X
   - Amélioration Y
   
   ### Modifié
   - Correction Z
   ```

4. **Créer un commit et un tag Git** :
   ```bash
   git add VERSION CHANGELOG.md frontend/package.json
   git commit -m "chore: bump version to 1.0.0"
   git tag -a v1.0.0 -m "Version 1.0.0"
   ```

5. **Pousser vers le dépôt** :
   ```bash
   git push origin main
   git push origin v1.0.0
   ```

## Workflow Recommandé

### Pour la Version 1.0.0 (Release Initiale)

1. **S'assurer que tout est prêt** :
   - Code testé et fonctionnel
   - Documentation à jour
   - CHANGELOG.md complété

2. **Créer la version** :
   ```powershell
   .\scripts\version.ps1 -Version 1.0.0
   ```

3. **Vérifier les changements** :
   ```bash
   git status
   git diff
   ```

4. **Créer le tag et pousser** :
   ```bash
   git push origin main
   git push origin v1.0.0
   ```

### Pour les Versions Suivantes (1.1.0, 1.2.0, etc.)

1. **Développer sur la branche `dev`**

2. **Quand prêt pour release** :
   ```bash
   git checkout main
   git merge dev
   ```

3. **Créer la nouvelle version** :
   ```powershell
   # Pour une nouvelle fonctionnalité (minor)
   .\scripts\version.ps1 -Type minor
   
   # Pour un correctif (patch)
   .\scripts\version.ps1 -Type patch
   ```

4. **Mettre à jour CHANGELOG.md** avec les changements

5. **Pousser** :
   ```bash
   git push origin main
   git push origin v1.1.0  # ou la version créée
   ```

## Exemples de Versions

- **1.0.0** : Release initiale
- **1.0.1** : Correction de bugs mineurs
- **1.1.0** : Nouvelles fonctionnalités (ex: Port-forward)
- **1.2.0** : Nouvelles fonctionnalités (ex: Multi-cluster)
- **2.0.0** : Changements majeurs incompatibles

## Tags Git

Les tags suivent le format `vX.Y.Z` :
- `v1.0.0`
- `v1.1.0`
- `v1.2.3`

Pour lister les tags :
```bash
git tag -l
git tag -l "v1.*"  # Versions 1.x
```

Pour voir les détails d'un tag :
```bash
git show v1.0.0
```

## GitHub Releases (Optionnel)

Si vous utilisez GitHub, vous pouvez créer des releases automatiquement via le workflow `.github/workflows/release.yml` qui se déclenche lors d'un push de tag.

## Notes

- Toujours mettre à jour `CHANGELOG.md` avant de créer une version
- Les versions doivent être créées sur la branche `main` (ou `master`)
- Ne jamais modifier un tag existant (créer une nouvelle version à la place)
