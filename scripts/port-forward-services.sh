#!/bin/bash

# Script pour faire le port-forward des services AWX et Ansible
# Usage: ./scripts/port-forward-services.sh

set -e

echo "🔌 Configuration du port-forward pour les services..."

# Vérifier que kubectl est configuré
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ kubectl n'est pas configuré ou le cluster n'est pas accessible"
    exit 1
fi

# Vérifier que les services existent
if ! kubectl get svc awx -n kura &> /dev/null; then
    echo "⚠️  Le service AWX n'existe pas dans le namespace kura"
    echo "   Déployez d'abord les services avec: kubectl apply -k infrastructure/k8s/"
    exit 1
fi

if ! kubectl get svc ansible-service -n kura &> /dev/null; then
    echo "⚠️  Le service ansible-service n'existe pas dans le namespace kura"
    echo "   Déployez d'abord les services avec: kubectl apply -k infrastructure/k8s/"
    exit 1
fi

# Fonction pour nettoyer les processus de port-forward existants
cleanup() {
    echo ""
    echo "🛑 Arrêt des port-forwards..."
    pkill -f "kubectl port-forward.*awx" || true
    pkill -f "kubectl port-forward.*ansible-service" || true
    exit 0
}

# Capturer Ctrl+C
trap cleanup SIGINT SIGTERM

# Démarrer les port-forwards en arrière-plan
echo "📡 Démarrage du port-forward pour AWX (port 8080)..."
kubectl port-forward svc/awx 8080:8080 -n kura > /dev/null 2>&1 &
AWX_PID=$!

echo "📡 Démarrage du port-forward pour le service Ansible (port 8083)..."
kubectl port-forward svc/ansible-service 8083:8083 -n kura > /dev/null 2>&1 &
ANSIBLE_PID=$!

# Attendre un peu pour vérifier que les port-forwards fonctionnent
sleep 2

if ps -p $AWX_PID > /dev/null && ps -p $ANSIBLE_PID > /dev/null; then
    echo ""
    echo "✅ Port-forwards actifs :"
    echo "   🌐 AWX:              http://localhost:8080"
    echo "   🔧 Service Ansible:  http://localhost:8083"
    echo "   📚 Documentation:    http://localhost:8083/docs"
    echo "   📊 Métriques:        http://localhost:8083/metrics"
    echo ""
    echo "💡 Dans GitHub Codespaces, exposez ces ports dans l'onglet 'Ports'"
    echo "   pour obtenir des URLs publiques."
    echo ""
    echo "Appuyez sur Ctrl+C pour arrêter les port-forwards"
    
    # Attendre indéfiniment
    wait
else
    echo "❌ Erreur lors du démarrage des port-forwards"
    cleanup
    exit 1
fi
