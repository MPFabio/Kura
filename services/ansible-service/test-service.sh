#!/bin/bash

# Script de test pour le service Ansible
# Usage: ./test-service.sh

set -e

SERVICE_URL="${SERVICE_URL:-http://localhost:8083}"
AWX_URL="${AWX_URL:-http://localhost:8080}"

echo "🧪 Test du service Ansible"
echo "Service URL: $SERVICE_URL"
echo "AWX URL: $AWX_URL"
echo ""

# Fonction pour tester un endpoint
test_endpoint() {
    local name=$1
    local method=$2
    local url=$3
    local data=$4
    
    echo "📋 Test: $name"
    if [ -n "$data" ]; then
        response=$(curl -s -X "$method" "$url" \
            -H "Content-Type: application/json" \
            -d "$data" \
            -w "\n%{http_code}")
    else
        response=$(curl -s -X "$method" "$url" -w "\n%{http_code}")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo "  ✅ Succès (HTTP $http_code)"
        if command -v jq &> /dev/null && [ -n "$body" ]; then
            echo "$body" | jq '.' 2>/dev/null | head -10 || echo "$body" | head -5
        fi
    else
        echo "  ❌ Échec (HTTP $http_code)"
        echo "$body" | head -5
    fi
    echo ""
}

# 1. Vérifier la santé du service
test_endpoint "Health check" "GET" "$SERVICE_URL/health"

# 2. Vérifier AWX
echo "📋 Test: Connexion AWX"
awx_response=$(curl -s -w "\n%{http_code}" "$AWX_URL/api/v2/ping/" || echo -e "\n000")
awx_code=$(echo "$awx_response" | tail -n1)
if [ "$awx_code" = "200" ]; then
    echo "  ✅ AWX accessible"
else
    echo "  ⚠️  AWX non accessible (HTTP $awx_code)"
    echo "     Assurez-vous qu'AWX est démarré sur $AWX_URL"
fi
echo ""

# 3. Tester les endpoints principaux
test_endpoint "Liste des jobs" "GET" "$SERVICE_URL/api/v1/ansible/jobs"
test_endpoint "Liste des inventaires" "GET" "$SERVICE_URL/api/v1/ansible/inventories"
test_endpoint "Liste des templates" "GET" "$SERVICE_URL/api/v1/ansible/job-templates"
test_endpoint "Liste des organisations" "GET" "$SERVICE_URL/api/v1/ansible/organizations"
test_endpoint "Liste des credentials" "GET" "$SERVICE_URL/api/v1/ansible/credentials"
test_endpoint "Liste des projets" "GET" "$SERVICE_URL/api/v1/ansible/projects"

# 4. Tester l'analyse de playbook
test_endpoint "Analyse de playbook" "POST" "$SERVICE_URL/api/v1/ansible/playbooks/analyze" '{
  "playbook_content": "---\n- name: Test playbook\n  hosts: localhost\n  tasks:\n    - name: Test task\n      debug:\n        msg: Hello World"
}'

# 5. Vérifier les métriques Prometheus
echo "📋 Test: Métriques Prometheus"
metrics=$(curl -s "$SERVICE_URL/metrics")
if echo "$metrics" | grep -q "ansible_api_requests_total"; then
    echo "  ✅ Métriques Prometheus disponibles"
    echo "$metrics" | grep "ansible_" | head -5
else
    echo "  ⚠️  Métriques Prometheus non trouvées"
fi
echo ""

echo "✅ Tests terminés"
echo ""
echo "💡 Pour tester les WebSockets, utilisez:"
echo "   wscat -c \"ws://localhost:8083/api/v1/ansible/jobs/1/stream\""
echo ""
echo "💡 Documentation interactive:"
echo "   Swagger UI: $SERVICE_URL/docs"
echo "   ReDoc: $SERVICE_URL/redoc"
