"""Point d'entrée principal du service Ansible."""
import logging
import sys
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from internal.config.config import load_config
from internal.cache.redis import RedisClient
from internal.client.tower_client import AnsibleTowerClient
from internal.service.ansible_service import AnsibleService
from internal.handler.ansible_handler import AnsibleHandler, create_router
from internal.handler.webhook_handler import WebhookHandler, create_webhook_router
from internal.handler.websocket_handler import WebSocketHandler
from internal.metrics.prometheus import get_metrics_response

# Configuration du logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[
        logging.StreamHandler(sys.stdout),
    ],
)

logger = logging.getLogger(__name__)

# Variables globales pour les ressources
config = None
cache = None
tower_client = None
ansible_service = None
webhook_handler = None
websocket_handler = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Gère le cycle de vie de l'application."""
    global config, cache, tower_client, ansible_service, webhook_handler, websocket_handler

    # Démarrage
    logger.info("Démarrage du service Ansible...")
    try:
        config = load_config()
        logger.info(f"Configuration chargée - Port: {config.server_port}, Env: {config.environment}")

        # Initialiser Redis
        try:
            cache = RedisClient(config)
            logger.info("Cache Redis initialisé")
        except Exception as e:
            logger.error(f"Erreur lors de l'initialisation de Redis: {e}")
            logger.warning("Le service continuera sans cache")
            cache = None

        # Initialiser le client Ansible Tower
        tower_client = AnsibleTowerClient(config)
        if config.ansible_tower_url:
            logger.info(f"Client Ansible Tower initialisé pour {config.ansible_tower_url}")
        else:
            logger.warning("ANSIBLE_TOWER_URL non configuré - certaines fonctionnalités seront limitées")

        # Initialiser le service métier (cache peut être None)
        ansible_service = AnsibleService(tower_client, cache, config)
        logger.info("Service Ansible initialisé")

        # Initialiser les handlers webhook et websocket
        try:
            webhook_handler = WebhookHandler()
            websocket_handler = WebSocketHandler(tower_client)
            logger.info("Handlers webhook et websocket initialisés")
        except Exception as e:
            logger.warning(f"Erreur lors de l'initialisation des handlers webhook/websocket: {e}")
            webhook_handler = None
            websocket_handler = None

    except Exception as e:
        logger.error(f"Erreur lors de l'initialisation: {e}", exc_info=True)
        # Ne pas lever l'exception pour permettre au service de démarrer même en mode dégradé
        logger.warning("Le service démarre en mode dégradé")

    # Configurer les routes maintenant que tout est initialisé (même en mode dégradé)
    logger.info("Configuration des routes...")
    setup_routes()  # setup_routes() gère déjà ses propres exceptions

    yield

    # Arrêt
    logger.info("Arrêt du service Ansible...")
    if cache:
        try:
            cache.close()
            logger.info("Cache Redis fermé")
        except Exception as e:
            logger.error(f"Erreur lors de la fermeture de Redis: {e}")


# Créer l'application FastAPI
app = FastAPI(
    title="Ansible Service API",
    description="Service REST pour interagir avec Ansible Tower",
    version="1.0.0",
    lifespan=lifespan,
)

# Middleware CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
@app.head("/health")
async def health_check():
    """Endpoint de santé."""
    try:
        return JSONResponse(
            content={
                "status": "ok",
                "service": "ansible-service",
                "ansible_tower_configured": config.ansible_tower_url is not None if config else False,
                "redis_available": cache is not None,
            }
        )
    except Exception as e:
        logger.error(f"Erreur dans health check: {e}")
        return JSONResponse(
            content={
                "status": "error",
                "service": "ansible-service",
                "error": str(e),
            },
            status_code=500
        )


@app.get("/metrics")
async def metrics():
    """Endpoint Prometheus pour les métriques."""
    return get_metrics_response()


# Initialiser les handlers et routes
def setup_routes():
    """Configure les routes de l'API."""
    global ansible_service, webhook_handler, websocket_handler, tower_client

    try:
        logger.info(f"Configuration des routes - ansible_service: {ansible_service is not None}, webhook_handler: {webhook_handler is not None}")
        
        if ansible_service is None:
            logger.warning("Service Ansible non initialisé - les routes ne seront pas disponibles")
            return

        # Routes principales Ansible
        try:
            logger.info("Création du handler Ansible...")
            handler = AnsibleHandler(ansible_service)
            logger.info("Handler Ansible créé avec succès")
            
            logger.info("Création du router Ansible...")
            router = create_router(handler)
            logger.info("Router Ansible créé avec succès")
            
            logger.info("Ajout du router Ansible à l'application...")
            app.include_router(router, prefix="/api/v1")
            logger.info("✅ Router Ansible ajouté avec succès")
        except Exception as e:
            logger.error(f"❌ Erreur lors de la configuration des routes Ansible: {e}", exc_info=True)
            import traceback
            logger.error(f"Traceback: {traceback.format_exc()}")
            # Continuer pour configurer les autres routes même si les routes Ansible échouent

        # Routes webhooks
        if webhook_handler:
            logger.info("Configuration des routes webhooks...")
            webhook_router = create_webhook_router(webhook_handler)
            app.include_router(webhook_router, prefix="/api/v1")
            logger.info("Routes webhooks ajoutées avec succès")
        else:
            logger.warning("webhook_handler est None - les routes webhooks ne seront pas disponibles")

        # Route WebSocket pour streaming
        if websocket_handler and tower_client:
            logger.info("Configuration de la route WebSocket...")
            from fastapi import WebSocket, Path
            
            async def websocket_endpoint(websocket: WebSocket, job_id: int = Path(...)):
                await websocket_handler.handle_websocket(websocket, job_id)
            
            app.add_api_websocket_route("/api/v1/ansible/jobs/{job_id}/stream", websocket_endpoint)
            logger.info("Route WebSocket ajoutée avec succès")
        else:
            logger.info("WebSocket non configuré (websocket_handler ou tower_client manquant)")
        
        logger.info("✅ Routes configurées avec succès")
    except Exception as e:
        logger.error(f"❌ Erreur lors de la configuration des routes: {e}", exc_info=True)
        # Ne pas lever l'exception pour permettre au service de démarrer même si certaines routes échouent
        import traceback
        logger.error(f"Traceback complet: {traceback.format_exc()}")


# Les routes sont maintenant configurées dans le lifespan avant le yield


if __name__ == "__main__":
    import uvicorn

    # Charger la config pour le port
    cfg = load_config()
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=int(cfg.server_port),
        log_level=cfg.log_level.lower(),
        reload=cfg.environment == "development",
    )
