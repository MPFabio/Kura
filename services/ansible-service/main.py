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
from internal.handler.websocket_handler import WebSocketHandler, create_websocket_route
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

        # Les routes seront configurées dans startup_event()
        logger.info("Initialisation terminée, configuration des routes...")

    except Exception as e:
        logger.error(f"Erreur lors de l'initialisation: {e}", exc_info=True)
        # Ne pas lever l'exception pour permettre au service de démarrer même en mode dégradé
        logger.warning("Le service démarre en mode dégradé")

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
        if ansible_service is None:
            logger.warning("Service Ansible non initialisé - les routes ne seront pas disponibles")
            return

        # Routes principales Ansible
        handler = AnsibleHandler(ansible_service)
        router = create_router(handler)
        app.include_router(router, prefix="/api/v1")

        # Routes webhooks
        if webhook_handler:
            webhook_router = create_webhook_router(webhook_handler)
            app.include_router(webhook_router, prefix="/api/v1")

        # Route WebSocket pour streaming
        if websocket_handler and tower_client:
            from fastapi import WebSocket, Path
            
            async def websocket_endpoint(websocket: WebSocket, job_id: int = Path(...)):
                await websocket_handler.handle_websocket(websocket, job_id)
            
            app.add_api_websocket_route("/api/v1/ansible/jobs/{job_id}/stream", websocket_endpoint)
        
        logger.info("Routes configurées avec succès")
    except Exception as e:
        logger.error(f"Erreur lors de la configuration des routes: {e}", exc_info=True)
        # Ne pas lever l'exception pour permettre au service de démarrer même si certaines routes échouent


# Configurer les routes après le démarrage
@app.on_event("startup")
async def startup_event():
    """Événement de démarrage."""
    try:
        setup_routes()
        logger.info("Service Ansible prêt et routes configurées")
    except Exception as e:
        logger.error(f"Erreur lors de la configuration des routes au démarrage: {e}", exc_info=True)
        # Le service peut continuer même si certaines routes échouent


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
