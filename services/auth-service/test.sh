#!/bin/bash

# Script de test pour le service d'authentification
# Ce script teste toutes les fonctionnalités du service

BASE_URL="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "🧪 Tests du service d'authentification"
echo "=========================================="
echo ""

# Test 1 : Vérifier la santé du service
echo "📋 Test 1 : Vérification de la santé du service"
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
BODY=$(echo "$HEALTH_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ Service accessible${NC}"
    echo "Réponse: $BODY"
else
    echo -e "${RED}❌ Service non accessible (code: $HTTP_CODE)${NC}"
    echo "Assurez-vous que le service est démarré : docker-compose up -d auth-service"
    exit 1
fi
echo ""

# Test 2 : Créer un utilisateur
echo "📋 Test 2 : Création d'un utilisateur"
REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test'$(date +%s)'@example.com",
    "username": "testuser'$(date +%s)'",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }')

HTTP_CODE=$(echo "$REGISTER_RESPONSE" | tail -n1)
BODY=$(echo "$REGISTER_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" -eq 201 ]; then
    echo -e "${GREEN}✅ Utilisateur créé avec succès${NC}"
    echo "Réponse: $BODY"
    USER_EMAIL=$(echo "$BODY" | grep -o '"email":"[^"]*"' | cut -d'"' -f4)
else
    echo -e "${YELLOW}⚠️  Code HTTP: $HTTP_CODE${NC}"
    echo "Réponse: $BODY"
    # Peut-être que l'utilisateur existe déjà, continuons avec un email fixe
    USER_EMAIL="test@example.com"
fi
echo ""

# Test 3 : Se connecter
echo "📋 Test 3 : Connexion"
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "'"$USER_EMAIL"'",
    "password": "password123"
  }')

HTTP_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
BODY=$(echo "$LOGIN_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" -eq 200 ]; then
    echo -e "${GREEN}✅ Connexion réussie${NC}"
    TOKEN=$(echo "$BODY" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    REFRESH_TOKEN=$(echo "$BODY" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
    echo "Token obtenu: ${TOKEN:0:50}..."
else
    echo -e "${RED}❌ Échec de la connexion (code: $HTTP_CODE)${NC}"
    echo "Réponse: $BODY"
    echo -e "${YELLOW}💡 Astuce: Créez d'abord un utilisateur avec le test 2${NC}"
    exit 1
fi
echo ""

# Test 4 : Accéder à une route protégée
if [ -n "$TOKEN" ]; then
    echo "📋 Test 4 : Accès à une route protégée (/auth/me)"
    ME_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$BASE_URL/api/v1/auth/me" \
      -H "Authorization: Bearer $TOKEN")
    
    HTTP_CODE=$(echo "$ME_RESPONSE" | tail -n1)
    BODY=$(echo "$ME_RESPONSE" | head -n-1)
    
    if [ "$HTTP_CODE" -eq 200 ]; then
        echo -e "${GREEN}✅ Route protégée accessible${NC}"
        echo "Réponse: $BODY"
    else
        echo -e "${RED}❌ Échec d'accès à la route protégée (code: $HTTP_CODE)${NC}"
        echo "Réponse: $BODY"
    fi
    echo ""
fi

# Test 5 : Rafraîchir le token
if [ -n "$REFRESH_TOKEN" ]; then
    echo "📋 Test 5 : Rafraîchissement du token"
    REFRESH_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/v1/auth/refresh" \
      -H "Content-Type: application/json" \
      -d '{
        "refresh_token": "'"$REFRESH_TOKEN"'"
      }')
    
    HTTP_CODE=$(echo "$REFRESH_RESPONSE" | tail -n1)
    BODY=$(echo "$REFRESH_RESPONSE" | head -n-1)
    
    if [ "$HTTP_CODE" -eq 200 ]; then
        echo -e "${GREEN}✅ Token rafraîchi avec succès${NC}"
        NEW_TOKEN=$(echo "$BODY" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo "Nouveau token obtenu: ${NEW_TOKEN:0:50}..."
    else
        echo -e "${YELLOW}⚠️  Échec du rafraîchissement (code: $HTTP_CODE)${NC}"
        echo "Réponse: $BODY"
    fi
    echo ""
fi

# Test 6 : Modifier le profil
if [ -n "$TOKEN" ]; then
    echo "📋 Test 6 : Modification du profil"
    UPDATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X PUT "$BASE_URL/api/v1/auth/me" \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "first_name": "Updated",
        "last_name": "Name"
      }')
    
    HTTP_CODE=$(echo "$UPDATE_RESPONSE" | tail -n1)
    BODY=$(echo "$UPDATE_RESPONSE" | head -n-1)
    
    if [ "$HTTP_CODE" -eq 200 ]; then
        echo -e "${GREEN}✅ Profil modifié avec succès${NC}"
        echo "Réponse: $BODY"
    else
        echo -e "${YELLOW}⚠️  Échec de la modification (code: $HTTP_CODE)${NC}"
        echo "Réponse: $BODY"
    fi
    echo ""
fi

echo "=========================================="
echo -e "${GREEN}✅ Tests terminés !${NC}"
echo "=========================================="
echo ""
echo "💡 Pour voir les logs du service :"
echo "   docker-compose logs -f auth-service"
echo ""
echo "💡 Pour vérifier la base de données :"
echo "   docker exec -it kura-postgres psql -U kura -d kura"
echo "   SELECT * FROM users;"
