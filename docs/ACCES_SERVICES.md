# Guide d'accès aux services Kura

Ce document explique comment accéder à chaque service de l'infrastructure Kura.

## Services accessibles via navigateur web

### Kong Gateway
- **URL** : http://localhost:8000
- **Description** : API Gateway principal
- **Test** : Ouvrir dans le navigateur, devrait afficher `{"message":"no Route matched with those values"}`

### Kong Admin API
- **URL** : http://localhost:8001
- **Description** : Interface d'administration de Kong
- **Test** : Ouvrir dans le navigateur, devrait afficher la configuration JSON de Kong

### Prometheus
- **URL** : http://localhost:9090
- **Description** : Interface web de Prometheus pour consulter les métriques
- **Test** : Ouvrir dans le navigateur, interface graphique complète

### Grafana
- **URL** : http://localhost:3000
- **Identifiants** : 
  - Username: `admin`
  - Password: `admin`
- **Description** : Interface de visualisation des métriques et dashboards
- **Test** : Ouvrir dans le navigateur et se connecter

## Services nécessitant des clients spécifiques

### PostgreSQL (port 5432)

PostgreSQL est une base de données relationnelle. Il ne peut pas être accédé via un navigateur web.

#### Option 1 : Via Docker (recommandé pour les tests rapides)
```bash
docker exec -it modulops-postgres psql -U modulops -d modulops
```

#### Option 2 : Client en ligne de commande (si psql est installé)
```bash
psql -h localhost -U modulops -d modulops
# Mot de passe: modulops
```

#### Option 3 : Clients graphiques
- **pgAdmin** : https://www.pgadmin.org/
- **DBeaver** : https://dbeaver.io/
- **DataGrip** : https://www.jetbrains.com/datagrip/

**Configuration pour les clients graphiques :**
- Host: `localhost`
- Port: `5432`
- Database: `modulops`
- Username: `modulops`
- Password: `modulops`

### Redis (port 6379)

Redis est un cache en mémoire. Il ne peut pas être accédé via un navigateur web.

#### Option 1 : Via Docker (recommandé)
```bash
docker exec -it modulops-redis redis-cli
```

#### Option 2 : Commandes Redis directes
```bash
# Test de connexion
docker exec modulops-redis redis-cli ping
# Devrait répondre: PONG

# Exemples de commandes
docker exec modulops-redis redis-cli set test "hello"
docker exec modulops-redis redis-cli get test
```

#### Option 3 : Clients graphiques
- **RedisInsight** : https://redis.com/redis-enterprise/redis-insight/
- **Another Redis Desktop Manager** : https://github.com/qishibo/AnotherRedisDesktopManager
- **Redis Commander** : https://github.com/joeferner/redis-commander

**Configuration pour les clients graphiques :**
- Host: `localhost`
- Port: `6379`
- Password: (aucun pour l'instant, mais recommandé en production)

### Kafka (port 9092)

Kafka est un message broker. Il ne peut pas être accédé via un navigateur web.

#### Option 1 : Via Docker (outils Kafka)
```bash
# Lister les topics
docker exec modulops-kafka kafka-topics --bootstrap-server localhost:9092 --list

# Créer un topic
docker exec modulops-kafka kafka-topics --bootstrap-server localhost:9092 --create --topic test-topic

# Consommer des messages
docker exec modulops-kafka kafka-console-consumer --bootstrap-server localhost:9092 --topic test-topic --from-beginning

# Produire des messages (dans un autre terminal)
docker exec -it modulops-kafka kafka-console-producer --bootstrap-server localhost:9092 --topic test-topic
```

#### Option 2 : Clients graphiques
- **Kafka UI** : https://github.com/provectus/kafka-ui
- **Kafdrop** : https://github.com/obsidiandynamics/kafdrop
- **Offset Explorer** : https://www.kafkatool.com/

**Configuration pour les clients graphiques :**
- Bootstrap Server: `localhost:9092`
- Zookeeper: `localhost:2181` (si nécessaire)

## Vérification rapide de tous les services

### Script de vérification

Créez un fichier `check-services.sh` :

```bash
#!/bin/bash

echo "Vérification des services Kura..."
echo ""

echo "Services HTTP (navigateur):"
echo "  - Kong Gateway: http://localhost:8000"
echo "  - Kong Admin: http://localhost:8001"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000"
echo ""

echo "Services base de données/cache:"
docker exec modulops-postgres pg_isready -U modulops && echo "  ✓ PostgreSQL: OK" || echo "  ✗ PostgreSQL: ERREUR"
docker exec modulops-redis redis-cli ping | grep -q PONG && echo "  ✓ Redis: OK" || echo "  ✗ Redis: ERREUR"
echo ""

echo "Services message broker:"
docker exec modulops-kafka kafka-broker-api-versions --bootstrap-server localhost:9092 > /dev/null 2>&1 && echo "  ✓ Kafka: OK" || echo "  ✗ Kafka: ERREUR"
docker exec modulops-zookeeper echo ruok | nc localhost 2181 | grep -q imok && echo "  ✓ Zookeeper: OK" || echo "  ✗ Zookeeper: ERREUR"
echo ""

echo "Services HTTP (curl):"
curl -s http://localhost:8001/ > /dev/null && echo "  ✓ Kong Admin: OK" || echo "  ✗ Kong Admin: ERREUR"
curl -s http://localhost:9090/-/healthy | grep -q Healthy && echo "  ✓ Prometheus: OK" || echo "  ✗ Prometheus: ERREUR"
curl -s http://localhost:3000/api/health > /dev/null && echo "  ✓ Grafana: OK" || echo "  ✗ Grafana: ERREUR"
```

## Résumé

| Service | Port | Type | Accès |
|---------|------|------|-------|
| Kong Gateway | 8000 | HTTP | Navigateur |
| Kong Admin | 8001 | HTTP | Navigateur |
| Prometheus | 9090 | HTTP | Navigateur |
| Grafana | 3000 | HTTP | Navigateur |
| PostgreSQL | 5432 | Base de données | Client SQL |
| Redis | 6379 | Cache | redis-cli ou client |
| Kafka | 9092 | Message broker | Client Kafka |
| Zookeeper | 2181 | Coordination | Client Zookeeper |

## Notes importantes

1. **PostgreSQL, Redis et Kafka ne sont PAS des serveurs web** - ils utilisent leurs propres protocoles
2. Pour accéder à ces services, vous devez utiliser des clients spécifiques ou des outils en ligne de commande
3. Les services HTTP (Kong, Prometheus, Grafana) sont accessibles directement via un navigateur
4. En production, tous les services devraient être sécurisés avec des mots de passe et des certificats SSL
