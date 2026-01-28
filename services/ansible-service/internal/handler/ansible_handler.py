"""Handlers HTTP pour le service Ansible."""
import logging
import time
from typing import Optional
from fastapi import APIRouter, HTTPException, Query, Path, Request
from fastapi.responses import JSONResponse

from internal.service.ansible_service import AnsibleService
from internal.models.models import (
    JobDetail,
    InventoryDetail,
    JobTemplateDetail,
    JobHistoryResponse,
)
from internal.metrics.prometheus import (
    api_requests_total,
    api_request_duration_seconds,
    jobs_launched_total,
    cache_hits_total,
    cache_misses_total,
)

logger = logging.getLogger(__name__)


class AnsibleHandler:
    """Handlers HTTP pour les endpoints Ansible."""

    def __init__(self, ansible_service: AnsibleService):
        """Initialise le handler."""
        self.service = ansible_service

    def _track_request(self, method: str, endpoint: str, status: int, duration: float):
        """Enregistre les métriques pour une requête."""
        api_requests_total.labels(method=method, endpoint=endpoint, status=status).inc()
        api_request_duration_seconds.labels(method=method, endpoint=endpoint).observe(duration)

    def get_jobs(
        self,
        page: int = Query(1, ge=1, description="Numéro de page"),
        page_size: int = Query(20, ge=1, le=100, description="Taille de la page"),
    ):
        """Récupère la liste des jobs."""
        try:
            result = self.service.get_jobs(page, page_size)
            if result is None:
                raise HTTPException(
                    status_code=503,
                    detail="Service Ansible Tower non disponible. Vérifiez la configuration."
                )
            return result
        except Exception as e:
            logger.error(f"Erreur lors de la récupération des jobs: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_job(
        self,
        job_id: int = Path(..., description="ID du job"),
        include_stdout: bool = Query(False, description="Inclure la sortie standard"),
    ):
        """Récupère les détails d'un job spécifique."""
        try:
            job = self.service.get_job(job_id, include_stdout)
            if job is None:
                raise HTTPException(status_code=404, detail=f"Job {job_id} non trouvé")
            return job
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération du job {job_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_inventories(
        self,
        page: int = Query(1, ge=1, description="Numéro de page"),
        page_size: int = Query(20, ge=1, le=100, description="Taille de la page"),
    ):
        """Récupère la liste des inventaires."""
        try:
            result = self.service.get_inventories(page, page_size)
            if result is None:
                raise HTTPException(
                    status_code=503,
                    detail="Service Ansible Tower non disponible. Vérifiez la configuration."
                )
            return result
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération des inventaires: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_inventory(
        self,
        inventory_id: int = Path(..., description="ID de l'inventaire"),
    ):
        """Récupère les détails d'un inventaire spécifique."""
        try:
            inventory = self.service.get_inventory(inventory_id)
            if inventory is None:
                raise HTTPException(status_code=404, detail=f"Inventaire {inventory_id} non trouvé")
            return inventory
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération de l'inventaire {inventory_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_inventory_hosts(
        self,
        inventory_id: int = Path(..., description="ID de l'inventaire"),
        page: int = Query(1, ge=1, description="Numéro de page"),
        page_size: int = Query(20, ge=1, le=100, description="Taille de la page"),
    ):
        """Récupère les hôtes d'un inventaire."""
        try:
            result = self.service.get_inventory_hosts(inventory_id, page, page_size)
            if result is None:
                raise HTTPException(status_code=404, detail=f"Inventaire {inventory_id} non trouvé")
            return result
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération des hôtes de l'inventaire {inventory_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_job_templates(
        self,
        page: int = Query(1, ge=1, description="Numéro de page"),
        page_size: int = Query(20, ge=1, le=100, description="Taille de la page"),
    ):
        """Récupère la liste des templates de jobs."""
        try:
            result = self.service.get_job_templates(page, page_size)
            if result is None:
                raise HTTPException(
                    status_code=503,
                    detail="Service Ansible Tower non disponible. Vérifiez la configuration."
                )
            return result
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération des templates de jobs: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_job_template(
        self,
        template_id: int = Path(..., description="ID du template"),
    ):
        """Récupère les détails d'un template de job spécifique."""
        try:
            template = self.service.get_job_template(template_id)
            if template is None:
                raise HTTPException(status_code=404, detail=f"Template {template_id} non trouvé")
            return template
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération du template {template_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def launch_job_template(
        self,
        template_id: int = Path(..., description="ID du template"),
        extra_vars: Optional[dict] = None,
    ):
        """Lance un job depuis un template."""
        start_time = time.time()
        try:
            job = self.service.launch_job_template(template_id, extra_vars)
            if job is None:
                raise HTTPException(status_code=400, detail=f"Impossible de lancer le template {template_id}")
            jobs_launched_total.labels(template_id=str(template_id)).inc()
            duration = time.time() - start_time
            self._track_request("POST", f"/job-templates/{template_id}/launch", 200, duration)
            return job
        except HTTPException as e:
            duration = time.time() - start_time
            self._track_request("POST", f"/job-templates/{template_id}/launch", e.status_code, duration)
            raise
        except Exception as e:
            duration = time.time() - start_time
            self._track_request("POST", f"/job-templates/{template_id}/launch", 500, duration)
            logger.error(f"Erreur lors du lancement du template {template_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_projects(
        self,
        page: int = Query(1, ge=1, description="Numéro de page"),
        page_size: int = Query(20, ge=1, le=100, description="Taille de la page"),
    ):
        """Récupère la liste des projets."""
        try:
            result = self.service.get_projects(page, page_size)
            if result is None:
                raise HTTPException(
                    status_code=503,
                    detail="Service Ansible Tower non disponible. Vérifiez la configuration."
                )
            return result
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération des projets: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_playbooks(
        self,
        project_id: int = Path(..., description="ID du projet"),
    ):
        """Récupère la liste des playbooks d'un projet."""
        try:
            playbooks = self.service.get_playbooks(project_id)
            return {"project_id": project_id, "playbooks": playbooks}
        except Exception as e:
            logger.error(f"Erreur lors de la récupération des playbooks du projet {project_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_job_history(
        self,
        template_id: Optional[int] = Query(None, description="ID du template (optionnel)"),
        page: int = Query(1, ge=1, description="Numéro de page"),
        page_size: int = Query(50, ge=1, le=100, description="Taille de la page"),
    ):
        """Récupère l'historique des jobs."""
        try:
            history = self.service.get_job_history(template_id, page, page_size)
            if history is None:
                raise HTTPException(
                    status_code=503,
                    detail="Service Ansible Tower non disponible. Vérifiez la configuration."
                )
            return history
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération de l'historique des jobs: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    # Credentials
    def get_credentials(
        self,
        page: int = Query(1, ge=1, description="Numéro de page"),
        page_size: int = Query(20, ge=1, le=100, description="Taille de la page"),
    ):
        """Récupère la liste des credentials."""
        try:
            result = self.service.get_credentials(page, page_size)
            if result is None:
                raise HTTPException(
                    status_code=503,
                    detail="Service Ansible Tower non disponible. Vérifiez la configuration."
                )
            return result
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération des credentials: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_credential(
        self,
        credential_id: int = Path(..., description="ID du credential"),
    ):
        """Récupère les détails d'un credential."""
        try:
            credential = self.service.get_credential(credential_id)
            if credential is None:
                raise HTTPException(status_code=404, detail=f"Credential {credential_id} non trouvé")
            return credential
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération du credential {credential_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def create_credential(
        self,
        credential_data: dict,
    ):
        """Crée un nouveau credential."""
        try:
            result = self.service.create_credential(credential_data)
            if result is None:
                raise HTTPException(status_code=400, detail="Impossible de créer le credential")
            return result
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la création du credential: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def update_credential(
        self,
        credential_id: int = Path(..., description="ID du credential"),
        credential_data: dict = None,
    ):
        """Met à jour un credential."""
        try:
            result = self.service.update_credential(credential_id, credential_data)
            if result is None:
                raise HTTPException(status_code=404, detail=f"Credential {credential_id} non trouvé")
            return result
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la mise à jour du credential {credential_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def delete_credential(
        self,
        credential_id: int = Path(..., description="ID du credential"),
    ):
        """Supprime un credential."""
        try:
            result = self.service.delete_credential(credential_id)
            if not result:
                raise HTTPException(status_code=404, detail=f"Credential {credential_id} non trouvé")
            return {"message": f"Credential {credential_id} supprimé avec succès"}
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la suppression du credential {credential_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    # Organizations
    def get_organizations(
        self,
        page: int = Query(1, ge=1, description="Numéro de page"),
        page_size: int = Query(20, ge=1, le=100, description="Taille de la page"),
    ):
        """Récupère la liste des organisations."""
        try:
            result = self.service.get_organizations(page, page_size)
            if result is None:
                raise HTTPException(
                    status_code=503,
                    detail="Service Ansible Tower non disponible. Vérifiez la configuration."
                )
            return result
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération des organisations: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def get_organization(
        self,
        organization_id: int = Path(..., description="ID de l'organisation"),
    ):
        """Récupère les détails d'une organisation."""
        try:
            org = self.service.get_organization(organization_id)
            if org is None:
                raise HTTPException(status_code=404, detail=f"Organisation {organization_id} non trouvée")
            return org
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la récupération de l'organisation {organization_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def create_organization(
        self,
        org_data: dict,
    ):
        """Crée une nouvelle organisation."""
        try:
            result = self.service.create_organization(org_data)
            if result is None:
                raise HTTPException(status_code=400, detail="Impossible de créer l'organisation")
            return result
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la création de l'organisation: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def update_organization(
        self,
        organization_id: int = Path(..., description="ID de l'organisation"),
        org_data: dict = None,
    ):
        """Met à jour une organisation."""
        try:
            result = self.service.update_organization(organization_id, org_data)
            if result is None:
                raise HTTPException(status_code=404, detail=f"Organisation {organization_id} non trouvée")
            return result
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la mise à jour de l'organisation {organization_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    def delete_organization(
        self,
        organization_id: int = Path(..., description="ID de l'organisation"),
    ):
        """Supprime une organisation."""
        try:
            result = self.service.delete_organization(organization_id)
            if not result:
                raise HTTPException(status_code=404, detail=f"Organisation {organization_id} non trouvée")
            return {"message": f"Organisation {organization_id} supprimée avec succès"}
        except HTTPException:
            raise
        except Exception as e:
            logger.error(f"Erreur lors de la suppression de l'organisation {organization_id}: {e}")
            raise HTTPException(status_code=500, detail=str(e))

    # Playbook analysis
    def analyze_playbook(
        self,
        playbook_content: str,
    ):
        """Analyse un playbook en profondeur."""
        try:
            result = self.service.analyze_playbook(playbook_content)
            return result
        except Exception as e:
            logger.error(f"Erreur lors de l'analyse du playbook: {e}")
            raise HTTPException(status_code=500, detail=str(e))


def create_router(handler: AnsibleHandler) -> APIRouter:
    """Crée le routeur FastAPI pour les endpoints Ansible."""
    router = APIRouter(prefix="/ansible", tags=["ansible"])

    # Jobs
    router.add_api_route(
        "/jobs",
        handler.get_jobs,
        methods=["GET"],
        summary="Liste des jobs",
        description="Récupère la liste paginée des jobs Ansible Tower",
    )
    router.add_api_route(
        "/jobs/{job_id}",
        handler.get_job,
        methods=["GET"],
        summary="Détails d'un job",
        description="Récupère les détails d'un job spécifique",
    )

    # Inventaires
    router.add_api_route(
        "/inventories",
        handler.get_inventories,
        methods=["GET"],
        summary="Liste des inventaires",
        description="Récupère la liste paginée des inventaires",
    )
    router.add_api_route(
        "/inventories/{inventory_id}",
        handler.get_inventory,
        methods=["GET"],
        summary="Détails d'un inventaire",
        description="Récupère les détails d'un inventaire spécifique",
    )
    router.add_api_route(
        "/inventories/{inventory_id}/hosts",
        handler.get_inventory_hosts,
        methods=["GET"],
        summary="Hôtes d'un inventaire",
        description="Récupère les hôtes d'un inventaire",
    )

    # Templates de jobs
    router.add_api_route(
        "/job-templates",
        handler.get_job_templates,
        methods=["GET"],
        summary="Liste des templates de jobs",
        description="Récupère la liste paginée des templates de jobs",
    )
    router.add_api_route(
        "/job-templates/{template_id}",
        handler.get_job_template,
        methods=["GET"],
        summary="Détails d'un template",
        description="Récupère les détails d'un template de job spécifique",
    )
    router.add_api_route(
        "/job-templates/{template_id}/launch",
        handler.launch_job_template,
        methods=["POST"],
        summary="Lancer un template",
        description="Lance un job depuis un template",
    )

    # Projets
    router.add_api_route(
        "/projects",
        handler.get_projects,
        methods=["GET"],
        summary="Liste des projets",
        description="Récupère la liste paginée des projets",
    )
    router.add_api_route(
        "/projects/{project_id}/playbooks",
        handler.get_playbooks,
        methods=["GET"],
        summary="Playbooks d'un projet",
        description="Récupère la liste des playbooks d'un projet",
    )

    # Historique
    router.add_api_route(
        "/jobs/history",
        handler.get_job_history,
        methods=["GET"],
        summary="Historique des jobs",
        description="Récupère l'historique des jobs avec filtrage optionnel par template",
    )

    # Credentials
    router.add_api_route(
        "/credentials",
        handler.get_credentials,
        methods=["GET"],
        summary="Liste des credentials",
        description="Récupère la liste paginée des credentials",
    )
    router.add_api_route(
        "/credentials",
        handler.create_credential,
        methods=["POST"],
        summary="Créer un credential",
        description="Crée un nouveau credential",
    )
    router.add_api_route(
        "/credentials/{credential_id}",
        handler.get_credential,
        methods=["GET"],
        summary="Détails d'un credential",
        description="Récupère les détails d'un credential spécifique",
    )
    router.add_api_route(
        "/credentials/{credential_id}",
        handler.update_credential,
        methods=["PUT"],
        summary="Mettre à jour un credential",
        description="Met à jour un credential existant",
    )
    router.add_api_route(
        "/credentials/{credential_id}",
        handler.delete_credential,
        methods=["DELETE"],
        summary="Supprimer un credential",
        description="Supprime un credential",
    )

    # Organizations
    router.add_api_route(
        "/organizations",
        handler.get_organizations,
        methods=["GET"],
        summary="Liste des organisations",
        description="Récupère la liste paginée des organisations",
    )
    router.add_api_route(
        "/organizations",
        handler.create_organization,
        methods=["POST"],
        summary="Créer une organisation",
        description="Crée une nouvelle organisation",
    )
    router.add_api_route(
        "/organizations/{organization_id}",
        handler.get_organization,
        methods=["GET"],
        summary="Détails d'une organisation",
        description="Récupère les détails d'une organisation spécifique",
    )
    router.add_api_route(
        "/organizations/{organization_id}",
        handler.update_organization,
        methods=["PUT"],
        summary="Mettre à jour une organisation",
        description="Met à jour une organisation existante",
    )
    router.add_api_route(
        "/organizations/{organization_id}",
        handler.delete_organization,
        methods=["DELETE"],
        summary="Supprimer une organisation",
        description="Supprime une organisation",
    )

    # Playbook analysis
    router.add_api_route(
        "/playbooks/analyze",
        handler.analyze_playbook,
        methods=["POST"],
        summary="Analyser un playbook",
        description="Analyse en profondeur un playbook YAML",
    )

    return router
