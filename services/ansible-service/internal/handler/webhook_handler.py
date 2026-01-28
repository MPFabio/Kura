"""Handler pour les webhooks Ansible Tower."""
import logging
from typing import Dict, Any
from fastapi import APIRouter, HTTPException, Request
from fastapi.responses import JSONResponse
from datetime import datetime

from internal.models.models import WebhookEvent
from internal.metrics.prometheus import webhooks_received_total

logger = logging.getLogger(__name__)


class WebhookHandler:
    """Handler pour recevoir et traiter les webhooks d'Ansible Tower."""

    def __init__(self):
        """Initialise le handler webhook."""
        self.event_handlers = {}

    def register_handler(self, event_type: str, handler_func):
        """Enregistre un handler pour un type d'événement."""
        self.event_handlers[event_type] = handler_func

    async def handle_webhook(self, request: Request):
        """Traite un webhook reçu d'Ansible Tower."""
        try:
            # Récupérer les données du webhook
            data = await request.json()
            headers = dict(request.headers)

            # Extraire les informations de l'événement
            event_type = data.get("event", "unknown")
            job_id = data.get("job_id") or data.get("id")
            status = data.get("status")
            timestamp = datetime.now()

            logger.info(f"Webhook reçu: event={event_type}, job_id={job_id}, status={status}")

            # Enregistrer la métrique
            webhooks_received_total.labels(event_type=event_type).inc()

            # Créer l'objet événement
            webhook_event = WebhookEvent(
                event=event_type,
                job_id=job_id,
                job_template_id=data.get("job_template_id"),
                status=status,
                timestamp=timestamp,
                data=data,
            )

            # Appeler le handler spécifique si enregistré
            if event_type in self.event_handlers:
                try:
                    await self.event_handlers[event_type](webhook_event)
                except Exception as e:
                    logger.error(f"Erreur dans le handler pour {event_type}: {e}")

            # Log l'événement
            logger.debug(f"Événement webhook traité: {webhook_event.model_dump()}")

            return JSONResponse(
                content={"status": "received", "event": event_type, "job_id": job_id},
                status_code=200,
            )

        except Exception as e:
            logger.error(f"Erreur lors du traitement du webhook: {e}")
            raise HTTPException(status_code=500, detail=str(e))


def create_webhook_router(handler: WebhookHandler) -> APIRouter:
    """Crée le routeur pour les webhooks."""
    router = APIRouter(prefix="/webhooks", tags=["webhooks"])

    router.add_api_route(
        "/ansible-tower",
        handler.handle_webhook,
        methods=["POST"],
        summary="Webhook Ansible Tower",
        description="Endpoint pour recevoir les webhooks d'Ansible Tower",
    )

    return router
