"""Handler WebSocket pour le streaming temps réel des logs."""
import logging
from fastapi import WebSocket, WebSocketDisconnect

from internal.client.tower_client import AnsibleTowerClient
from internal.metrics.prometheus import active_websocket_connections

logger = logging.getLogger(__name__)


class WebSocketHandler:
    """Handler WebSocket pour le streaming des logs de jobs."""

    def __init__(self, tower_client: AnsibleTowerClient):
        """Initialise le handler WebSocket."""
        self.tower_client = tower_client
        self.active_connections: dict[int, WebSocket] = {}

    async def connect(self, websocket: WebSocket, job_id: int):
        """Accepte une connexion WebSocket pour un job."""
        await websocket.accept()
        self.active_connections[job_id] = websocket
        active_websocket_connections.inc()
        logger.info(f"Connexion WebSocket établie pour le job {job_id}")

    def disconnect(self, job_id: int):
        """Ferme la connexion WebSocket pour un job."""
        if job_id in self.active_connections:
            del self.active_connections[job_id]
            active_websocket_connections.dec()
            logger.info(f"Connexion WebSocket fermée pour le job {job_id}")

    async def stream_job_logs(self, websocket: WebSocket, job_id: int):
        """Stream les logs d'un job en temps réel."""
        try:
            await self.connect(websocket, job_id)

            # Récupérer le statut initial du job
            job = self.tower_client.get_job(job_id)
            if job:
                await websocket.send_json({
                    "type": "job_status",
                    "job_id": job_id,
                    "status": job.get("status"),
                    "data": job,
                })

            # Polling pour les mises à jour (simulation de streaming)
            # Dans une implémentation réelle, on utiliserait les webhooks ou l'API de streaming
            last_status = job.get("status") if job else None
            last_stdout_length = 0

            while True:
                try:
                    # Vérifier le statut du job
                    current_job = self.tower_client.get_job(job_id)
                    if current_job:
                        current_status = current_job.get("status")
                        if current_status != last_status:
                            await websocket.send_json({
                                "type": "status_update",
                                "job_id": job_id,
                                "status": current_status,
                                "previous_status": last_status,
                            })
                            last_status = current_status

                        # Si le job est terminé, envoyer les logs finaux et fermer
                        if current_status in ["successful", "failed", "error", "canceled"]:
                            stdout = self.tower_client.get_job_stdout(job_id)
                            if stdout and len(stdout) > last_stdout_length:
                                await websocket.send_json({
                                    "type": "stdout",
                                    "job_id": job_id,
                                    "content": stdout[last_stdout_length:],
                                })
                                last_stdout_length = len(stdout)

                            await websocket.send_json({
                                "type": "job_complete",
                                "job_id": job_id,
                                "status": current_status,
                                "stdout": stdout,
                            })
                            break

                        # Envoyer les nouveaux logs si disponibles
                        stdout = self.tower_client.get_job_stdout(job_id)
                        if stdout and len(stdout) > last_stdout_length:
                            await websocket.send_json({
                                "type": "stdout",
                                "job_id": job_id,
                                "content": stdout[last_stdout_length:],
                            })
                            last_stdout_length = len(stdout)

                    # Attendre avant le prochain polling
                    import asyncio
                    await asyncio.sleep(2)  # Poll toutes les 2 secondes

                except WebSocketDisconnect:
                    logger.info(f"Client déconnecté pour le job {job_id}")
                    break
                except Exception as e:
                    logger.error(f"Erreur lors du streaming pour le job {job_id}: {e}")
                    await websocket.send_json({
                        "type": "error",
                        "job_id": job_id,
                        "message": str(e),
                    })
                    break

        except WebSocketDisconnect:
            logger.info(f"WebSocket déconnecté pour le job {job_id}")
        except Exception as e:
            logger.error(f"Erreur dans stream_job_logs pour le job {job_id}: {e}")
        finally:
            self.disconnect(job_id)

    async def handle_websocket(self, websocket: WebSocket, job_id: int):
        """Gère une connexion WebSocket."""
        await self.stream_job_logs(websocket, job_id)


# La route WebSocket est créée directement dans main.py
