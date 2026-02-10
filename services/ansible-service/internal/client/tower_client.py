"""Client REST pour l'API Ansible Tower."""
import logging
from typing import Optional, Dict, Any, List
import httpx
from httpx import Timeout

from internal.config.config import Config

logger = logging.getLogger(__name__)


class AnsibleTowerClient:
    """Client pour interagir avec l'API REST d'Ansible Tower."""

    def __init__(self, config: Config):
        """Initialise le client Ansible Tower."""
        self.config = config
        self.base_url = config.ansible_tower_url.rstrip("/") if config.ansible_tower_url else None
        self.username = config.ansible_tower_username
        self.password = config.ansible_tower_password
        self.verify_ssl = config.ansible_tower_verify_ssl

        if not self.base_url:
            logger.warning("ANSIBLE_TOWER_URL non configuré - le client ne pourra pas fonctionner")

        self.timeout = Timeout(30.0, connect=10.0)
        self._auth = (
            httpx.BasicAuth(self.username, self.password)
            if self.username and self.password
            else None
        )

    def _get_headers(self) -> Dict[str, str]:
        """Retourne les en-têtes HTTP de base."""
        return {"Content-Type": "application/json"}

    def _request(self, method: str, endpoint: str, **kwargs) -> Optional[Dict[str, Any]]:
        """Effectue une requête HTTP vers l'API Ansible Tower."""
        if not self.base_url:
            logger.error("ANSIBLE_TOWER_URL non configuré")
            return None

        url = f"{self.base_url}/api/v2{endpoint}"
        headers = self._get_headers()

        try:
            with httpx.Client(timeout=self.timeout, verify=self.verify_ssl) as client:
                response = client.request(
                    method, url, headers=headers, auth=self._auth, **kwargs
                )
                response.raise_for_status()
                return response.json()
        except httpx.HTTPStatusError as e:
            logger.error(f"Erreur HTTP {e.response.status_code} pour {url}: {e.response.text}")
            return None
        except httpx.HTTPError as e:
            logger.error(f"Erreur lors de la requête vers {url}: {e}")
            return None

    def get_jobs(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des jobs (plus récents en premier)."""
        return self._request("GET", f"/jobs/?page={page}&page_size={page_size}&order_by=-finished")

    def get_job(self, job_id: int) -> Optional[Dict[str, Any]]:
        """Récupère les détails d'un job spécifique."""
        return self._request("GET", f"/jobs/{job_id}/")

    def get_job_stdout(self, job_id: int) -> Optional[str]:
        """Récupère la sortie standard d'un job."""
        if not self.base_url:
            return None

        url = f"{self.base_url}/api/v2/jobs/{job_id}/stdout/"
        headers = self._get_headers()

        try:
            with httpx.Client(timeout=self.timeout, verify=self.verify_ssl) as client:
                response = client.get(
                    url, headers=headers, auth=self._auth, params={"format": "txt"}
                )
                response.raise_for_status()
                return response.text
        except httpx.HTTPError as e:
            logger.error(f"Erreur lors de la récupération de la sortie du job {job_id}: {e}")
            return None

    def get_inventories(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des inventaires."""
        return self._request("GET", f"/inventories/?page={page}&page_size={page_size}")

    def get_inventory(self, inventory_id: int) -> Optional[Dict[str, Any]]:
        """Récupère les détails d'un inventaire spécifique."""
        return self._request("GET", f"/inventories/{inventory_id}/")

    def get_inventory_hosts(self, inventory_id: int, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère les hôtes d'un inventaire."""
        return self._request("GET", f"/inventories/{inventory_id}/hosts/?page={page}&page_size={page_size}")

    def get_job_templates(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des templates de jobs."""
        return self._request("GET", f"/job_templates/?page={page}&page_size={page_size}")

    def get_job_template(self, template_id: int) -> Optional[Dict[str, Any]]:
        """Récupère les détails d'un template de job spécifique."""
        return self._request("GET", f"/job_templates/{template_id}/")

    def launch_job_template(self, template_id: int, extra_vars: Optional[Dict[str, Any]] = None) -> Optional[Dict[str, Any]]:
        """Lance un job depuis un template."""
        data = {}
        if extra_vars:
            import json
            data["extra_vars"] = json.dumps(extra_vars)
        return self._request("POST", f"/job_templates/{template_id}/launch/", json=data)

    def get_project(self, project_id: int) -> Optional[Dict[str, Any]]:
        """Récupère les détails d'un projet."""
        return self._request("GET", f"/projects/{project_id}/")

    def get_projects(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des projets."""
        return self._request("GET", f"/projects/?page={page}&page_size={page_size}")

    def get_playbooks(self, project_id: int) -> Optional[List[Dict[str, Any]]]:
        """Récupère la liste des playbooks d'un projet."""
        project = self.get_project(project_id)
        if not project:
            return None

        # Note: Ansible Tower stocke les playbooks dans le projet
        # Cette méthode retourne les métadonnées du projet qui contient les playbooks
        playbooks = []
        if project.get("scm_type") == "git":
            # Les playbooks sont dans le dépôt Git
            # Pour une implémentation complète, il faudrait cloner le dépôt et parser les fichiers
            playbooks.append({
                "name": project.get("name", "unknown"),
                "scm_type": project.get("scm_type"),
                "scm_url": project.get("scm_url"),
            })
        return playbooks

    def get_job_history(self, template_id: Optional[int] = None, page: int = 1, page_size: int = 50) -> Optional[Dict[str, Any]]:
        """Récupère l'historique des jobs."""
        endpoint = "/jobs/"
        params = f"?page={page}&page_size={page_size}"
        if template_id:
            params += f"&job_template={template_id}"
        params += "&order_by=-finished"
        return self._request("GET", endpoint + params)

    # Credentials
    def get_credentials(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des credentials."""
        return self._request("GET", f"/credentials/?page={page}&page_size={page_size}")

    def get_credential(self, credential_id: int) -> Optional[Dict[str, Any]]:
        """Récupère les détails d'un credential."""
        return self._request("GET", f"/credentials/{credential_id}/")

    def create_credential(self, data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Crée un nouveau credential."""
        return self._request("POST", "/credentials/", json=data)

    def update_credential(self, credential_id: int, data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Met à jour un credential."""
        return self._request("PUT", f"/credentials/{credential_id}/", json=data)

    def delete_credential(self, credential_id: int) -> bool:
        """Supprime un credential."""
        result = self._request("DELETE", f"/credentials/{credential_id}/")
        return result is not None

    # Organizations
    def get_organizations(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des organisations."""
        return self._request("GET", f"/organizations/?page={page}&page_size={page_size}")

    def get_organization(self, organization_id: int) -> Optional[Dict[str, Any]]:
        """Récupère les détails d'une organisation."""
        return self._request("GET", f"/organizations/{organization_id}/")

    def create_organization(self, data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Crée une nouvelle organisation."""
        return self._request("POST", "/organizations/", json=data)

    def update_organization(self, organization_id: int, data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Met à jour une organisation."""
        return self._request("PUT", f"/organizations/{organization_id}/", json=data)

    def delete_organization(self, organization_id: int) -> bool:
        """Supprime une organisation."""
        result = self._request("DELETE", f"/organizations/{organization_id}/")
        return result is not None

    def get_project_playbook_content(self, project_id: int, playbook_path: str) -> Optional[str]:
        """Récupère le contenu d'un playbook depuis un projet."""
        # Note: Cette méthode nécessite que le projet soit synchronisé dans Tower
        # Pour une implémentation complète, il faudrait utiliser l'API de fichiers du projet
        # ou cloner le dépôt Git directement
        project = self.get_project(project_id)
        if not project:
            return None
        
        # Ansible Tower expose les fichiers via l'API /api/v2/projects/{id}/playbooks/
        # Mais pour obtenir le contenu, il faut généralement accéder au dépôt SCM
        # Cette méthode retourne None pour indiquer que le contenu n'est pas directement accessible
        # via l'API standard - il faudrait implémenter un clone Git ou utiliser l'API de fichiers
        return None
