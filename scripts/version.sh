#!/bin/bash

# Script de gestion des versions pour Kura
# Usage: ./scripts/version.sh [major|minor|patch] ou ./scripts/version.sh [version] (ex: 1.2.3)

set -e

VERSION_FILE="VERSION"
CHANGELOG_FILE="CHANGELOG.md"

# Lire la version actuelle
if [ -f "$VERSION_FILE" ]; then
    CURRENT_VERSION=$(cat "$VERSION_FILE")
else
    CURRENT_VERSION="0.0.0"
fi

# Fonction pour incrémenter la version
increment_version() {
    local version=$1
    local part=$2
    
    IFS='.' read -ra VERSION_PARTS <<< "$version"
    MAJOR=${VERSION_PARTS[0]}
    MINOR=${VERSION_PARTS[1]}
    PATCH=${VERSION_PARTS[2]}
    
    case $part in
        major)
            MAJOR=$((MAJOR + 1))
            MINOR=0
            PATCH=0
            ;;
        minor)
            MINOR=$((MINOR + 1))
            PATCH=0
            ;;
        patch)
            PATCH=$((PATCH + 1))
            ;;
        *)
            echo "Usage: $0 [major|minor|patch] ou $0 [version]"
            exit 1
            ;;
    esac
    
    echo "$MAJOR.$MINOR.$PATCH"
}

# Déterminer la nouvelle version
if [ -z "$1" ]; then
    echo "Version actuelle: $CURRENT_VERSION"
    echo "Usage: $0 [major|minor|patch] ou $0 [version]"
    exit 1
fi

# Vérifier si c'est une version spécifique (format X.Y.Z)
if [[ $1 =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    NEW_VERSION=$1
else
    NEW_VERSION=$(increment_version "$CURRENT_VERSION" "$1")
fi

echo "Version actuelle: $CURRENT_VERSION"
echo "Nouvelle version: $NEW_VERSION"

# Mettre à jour le fichier VERSION
echo "$NEW_VERSION" > "$VERSION_FILE"
echo "✓ Fichier VERSION mis à jour"

# Mettre à jour package.json du frontend
if [ -f "frontend/package.json" ]; then
    if command -v jq &> /dev/null; then
        jq ".version = \"$NEW_VERSION\"" frontend/package.json > frontend/package.json.tmp && mv frontend/package.json.tmp frontend/package.json
        echo "✓ frontend/package.json mis à jour"
    else
        echo "⚠ jq non installé, veuillez mettre à jour frontend/package.json manuellement"
    fi
fi

# Créer un tag Git
read -p "Créer un tag Git v$NEW_VERSION ? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    git add "$VERSION_FILE" "$CHANGELOG_FILE"
    if [ -f "frontend/package.json" ]; then
        git add frontend/package.json
    fi
    git commit -m "chore: bump version to $NEW_VERSION" || true
    git tag -a "v$NEW_VERSION" -m "Version $NEW_VERSION"
    echo "✓ Tag Git v$NEW_VERSION créé"
    echo ""
    echo "Pour pousser le tag vers le dépôt distant:"
    echo "  git push origin v$NEW_VERSION"
    echo "  git push origin main  # ou votre branche"
fi

echo ""
echo "Version $NEW_VERSION prête !"
echo "N'oubliez pas de mettre à jour CHANGELOG.md avec les changements de cette version."
