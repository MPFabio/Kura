#!/bin/bash

# Script de configuration de l'infrastructure Kubernetes pour Kura
# Ce script configure un cluster Kubernetes local et déploie tous les services

set -e

echo "Configuration de l'infrastructure Kura sur Kubernetes"

# Vérifier si kubectl est installé
if ! command -v kubectl &> /dev/null; then
    echo "kubectl n'est pas installé. Veuillez l'installer d'abord."
    exit 1
fi

# Vérifier si minikube ou kind est disponible
if command -v minikube &> /dev/null; then
    echo "Utilisation de minikube"
    CLUSTER_TYPE="minikube"
    if ! minikube status &> /dev/null; then
        echo "Démarrage de minikube..."
        minikube start --cpus=4 --memory=8192 --disk-size=20g
    fi
elif command -v kind &> /dev/null; then
    echo "Utilisation de kind"
    CLUSTER_TYPE="kind"
    if ! kind get clusters | grep -q kura; then
        echo "Création du cluster kind..."
        kind create cluster --name kura --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
    fi
else
    echo "Ni minikube ni kind n'est installé. Veuillez installer l'un d'eux."
    exit 1
fi

# Appliquer les manifests Kubernetes
echo "Application des manifests Kubernetes..."
kubectl apply -k infrastructure/k8s/

# Attendre que les services soient prêts
echo "⏳ Attente du déploiement des services..."
kubectl wait --for=condition=available --timeout=300s deployment/postgres -n kura || true
kubectl wait --for=condition=available --timeout=300s deployment/redis -n kura || true
kubectl wait --for=condition=ready --timeout=300s pod -l app=zookeeper -n kura || true
kubectl wait --for=condition=ready --timeout=300s pod -l app=kafka -n kura || true
kubectl wait --for=condition=available --timeout=300s deployment/kong -n kura || true
kubectl wait --for=condition=available --timeout=300s deployment/prometheus -n kura || true
kubectl wait --for=condition=available --timeout=300s deployment/grafana -n kura || true

echo "Infrastructure déployée avec succès!"
echo ""
echo "Services disponibles:"
echo "  - PostgreSQL: kubectl port-forward svc/postgres 5432:5432 -n kura"
echo "  - Redis: kubectl port-forward svc/redis 6379:6379 -n kura"
echo "  - Kafka: kubectl port-forward svc/kafka 9092:9092 -n kura"
echo "  - Kong Gateway: kubectl port-forward svc/kong 8000:8000 -n kura"
echo "  - Kong Admin: kubectl port-forward svc/kong 8001:8001 -n kura"
echo "  - Prometheus: kubectl port-forward svc/prometheus 9090:9090 -n kura"
echo "  - Grafana: kubectl port-forward svc/grafana 3000:3000 -n kura"
echo ""
echo "Vérifier le statut: kubectl get all -n kura"
