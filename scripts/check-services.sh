#!/bin/bash

# Script de vérification des services ModulOps

echo "🔍 Vérification des services ModulOps..."
echo ""

echo "📊 Services HTTP (accessibles via navigateur):"
echo "  - Kong Gateway: http://localhost:8000"
echo "  - Kong Admin: http://localhost:8001"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000"
echo ""

echo "🔌 Services base de données/cache (nécessitent des clients spécifiques):"
if docker exec modulops-postgres pg_isready -U modulops > /dev/null 2>&1; then
    echo "  ✅ PostgreSQL (port 5432): OK - Utilisez 'docker exec -it modulops-postgres psql -U modulops -d modulops'"
else
    echo "  ❌ PostgreSQL: ERREUR"
fi

if docker exec modulops-redis redis-cli ping 2>/dev/null | grep -q PONG; then
    echo "  ✅ Redis (port 6379): OK - Utilisez 'docker exec -it modulops-redis redis-cli'"
else
    echo "  ❌ Redis: ERREUR"
fi
echo ""

echo "📨 Services message broker:"
if docker exec modulops-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1; then
    echo "  ✅ Kafka (port 9092): OK - Utilisez un client Kafka"
else
    echo "  ❌ Kafka: ERREUR"
fi

if docker exec modulops-zookeeper echo ruok 2>/dev/null | nc localhost 2181 2>/dev/null | grep -q imok; then
    echo "  ✅ Zookeeper (port 2181): OK"
else
    echo "  ❌ Zookeeper: ERREUR"
fi
echo ""

echo "🌐 Services HTTP (test avec curl):"
if curl -s http://localhost:8001/ > /dev/null 2>&1; then
    echo "  ✅ Kong Admin (port 8001): OK - http://localhost:8001"
else
    echo "  ❌ Kong Admin: ERREUR"
fi

if curl -s http://localhost:9090/-/healthy 2>/dev/null | grep -q Healthy; then
    echo "  ✅ Prometheus (port 9090): OK - http://localhost:9090"
else
    echo "  ❌ Prometheus: ERREUR"
fi

if curl -s http://localhost:3000/api/health > /dev/null 2>&1; then
    echo "  ✅ Grafana (port 3000): OK - http://localhost:3000"
else
    echo "  ❌ Grafana: ERREUR"
fi
echo ""

echo "📝 Note: PostgreSQL, Redis et Kafka ne sont PAS accessibles via un navigateur web."
echo "   Consultez docs/ACCES_SERVICES.md pour savoir comment y accéder."
