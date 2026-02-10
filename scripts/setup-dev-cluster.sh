#!/bin/bash
# Script pour créer un cluster Kubernetes de dev dans le Codespace/local

set -e

echo "🔍 Vérification des outils disponibles..."

# Vérifier si k3d est installé
if ! command -v k3d &> /dev/null; then
    echo "📦 Installation de k3d..."
    curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
fi

# Vérifier si un cluster existe déjà
if k3d cluster list | grep -q "kura-dev"; then
    echo "✅ Cluster kura-dev existe déjà"
    echo "📋 Récupération du kubeconfig..."
    k3d kubeconfig get kura-dev
else
    echo "🚀 Création du cluster kura-dev..."
    k3d cluster create kura-dev \
        --api-port 0.0.0.0:6443 \
        --servers 1 \
        --agents 0 \
        --wait
    
    echo "✅ Cluster créé !"
    echo ""
    echo "📋 Kubeconfig (copie-le dans l'app Kura) :"
    echo "---"
    k3d kubeconfig get kura-dev
    echo "---"
fi
