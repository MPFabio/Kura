.PHONY: help setup-local setup-k8s up down logs clean status check

help: ## Affiche cette aide
	@echo "Commandes disponibles:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup-local: ## Configure et démarre l'infrastructure locale avec Docker Compose
	@echo "🚀 Configuration de l'infrastructure locale..."
	@bash scripts/setup-local.sh

setup-k8s: ## Configure et déploie l'infrastructure sur Kubernetes
	@echo "🚀 Configuration de l'infrastructure Kubernetes..."
	@bash scripts/setup-k8s.sh

up: ## Démarre tous les services avec Docker Compose
	@echo "📦 Démarrage des services..."
	docker-compose up -d

down: ## Arrête tous les services Docker Compose
	@echo "🛑 Arrêt des services..."
	docker-compose down

logs: ## Affiche les logs de tous les services
	docker-compose logs -f

logs-postgres: ## Affiche les logs de PostgreSQL
	docker-compose logs -f postgres

logs-redis: ## Affiche les logs de Redis
	docker-compose logs -f redis

logs-kafka: ## Affiche les logs de Kafka
	docker-compose logs -f kafka

logs-kong: ## Affiche les logs de Kong
	docker-compose logs -f kong

logs-prometheus: ## Affiche les logs de Prometheus
	docker-compose logs -f prometheus

logs-grafana: ## Affiche les logs de Grafana
	docker-compose logs -f grafana

status: ## Affiche le statut des services
	@echo "📊 Statut des services:"
	@docker-compose ps

check: ## Vérifie l'accessibilité de tous les services
	@bash scripts/check-services.sh

status-k8s: ## Affiche le statut des services sur Kubernetes
	@echo "📊 Statut des services Kubernetes:"
	@kubectl get all -n modulops

clean: ## Nettoie les volumes et données Docker
	@echo "🧹 Nettoyage des volumes..."
	docker-compose down -v
	@echo "✅ Nettoyage terminé"

restart: ## Redémarre tous les services
	@echo "🔄 Redémarrage des services..."
	docker-compose restart

restart-service: ## Redémarre un service spécifique (usage: make restart-service SERVICE=postgres)
	@if [ -z "$(SERVICE)" ]; then \
		echo "❌ Veuillez spécifier un service: make restart-service SERVICE=postgres"; \
	else \
		echo "🔄 Redémarrage de $(SERVICE)..."; \
		docker-compose restart $(SERVICE); \
	fi

port-forward-k8s: ## Configure les port-forwards pour Kubernetes
	@echo "🔌 Configuration des port-forwards..."
	@echo "PostgreSQL: kubectl port-forward svc/postgres 5432:5432 -n modulops"
	@echo "Redis: kubectl port-forward svc/redis 6379:6379 -n modulops"
	@echo "Kafka: kubectl port-forward svc/kafka 9092:9092 -n modulops"
	@echo "Kong: kubectl port-forward svc/kong 8000:8000 -n modulops"
	@echo "Prometheus: kubectl port-forward svc/prometheus 9090:9090 -n modulops"
	@echo "Grafana: kubectl port-forward svc/grafana 3000:3000 -n modulops"

apply-k8s: ## Applique les manifests Kubernetes
	@echo "📋 Application des manifests Kubernetes..."
	kubectl apply -k infrastructure/k8s/

delete-k8s: ## Supprime les ressources Kubernetes
	@echo "🗑️  Suppression des ressources Kubernetes..."
	kubectl delete -k infrastructure/k8s/

frontend-install: ## Installe les dépendances du frontend
	@echo "📦 Installation des dépendances du frontend..."
	cd frontend && npm install

frontend-dev: ## Démarre le serveur de développement du frontend
	@echo "🚀 Démarrage du frontend en mode développement..."
	cd frontend && npm run dev

frontend-build: ## Compile le frontend pour la production
	@echo "🏗️  Compilation du frontend..."
	cd frontend && npm run build

frontend-preview: ## Prévisualise le build de production du frontend
	@echo "👀 Prévisualisation du build..."
	cd frontend && npm run preview
