#!/bin/bash

# Script de configuration de l'infrastructure locale avec Docker Compose pour ModulOps

set -e

echo "Configuration de l'infrastructure ModulOps locale avec Docker Compose"

# Vérifier si Docker est installé
if ! command -v docker &> /dev/null; then
    echo "Docker n'est pas installé. Veuillez l'installer d'abord."
    exit 1
fi

# Vérifier si Docker Compose est installé
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo "Docker Compose n'est pas installé. Veuillez l'installer d'abord."
    exit 1
fi

# Utiliser docker compose (nouvelle syntaxe) ou docker-compose (ancienne syntaxe)
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

# Démarrer les services
echo "Démarrage des services..."
$DOCKER_COMPOSE up -d

# Attendre que les services soient prêts
echo "Attente du démarrage des services..."
sleep 10

# Vérifier la santé des services
echo "Vérification de la santé des services..."
$DOCKER_COMPOSE ps

echo "Infrastructure locale démarrée avec succès!"
echo ""
echo "Services disponibles:"
echo "  - PostgreSQL: localhost:5432 (user: modulops, password: modulops, db: modulops)"
echo "  - Redis: localhost:6379"
echo "  - Kafka: localhost:9092"
echo "  - Zookeeper: localhost:2181"
echo "  - Kong Gateway: http://localhost:8000"
echo "  - Kong Admin: http://localhost:8001"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000 (user: admin, password: admin)"
echo ""
echo "Pour arrêter les services: docker-compose down"
echo "Pour voir les logs: docker-compose logs -f [service]"
