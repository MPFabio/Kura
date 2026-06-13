"""Client REST pour l'API Ansible Semaphore.

Implémente la même interface que AnsibleTowerClient pour permettre
un switch transparent entre Tower/AWX et Semaphore.
"""
import logging
from typing import Optional, Dict, Any, List
from datetime import datetime, timezone
import httpx
from httpx import Timeout

from internal.config.config import Config

logger = logging.getLogger(__name__)

# Mapping statuts Semaphore → AWX pour compatibilité avec le frontend
STATUS_MAP = {
    "waiting":  "pending",
    "running":  "running",
    "success":  "successful",
    "error":    "failed",
    "stopped":  "canceled",
}


class SemaphoreClient:
    """Client pour interagir avec l'API REST d'Ansible Semaphore.

    Semaphore expose tout sous /api/projects/{project_id}/...
    On traduit les réponses au format AWX pour que le reste du service
    ne nécessite aucun changement.
    """

    def __init__(self, config: Config):
        self.base_url = config.semaphore_url.rstrip("/") if config.semaphore_url else None
        self.api_token = config.semaphore_api_token
        self.project_id = config.semaphore_project_id
        self.timeout = Timeout(30.0, connect=10.0)

        if not self.base_url:
            logger.warning("SEMAPHORE_URL non configuré")

    def _headers(self) -> Dict[str, str]:
        h = {"Content-Type": "application/json", "Accept": "application/json"}
        if self.api_token:
            h["Authorization"] = f"Bearer {self.api_token}"
        return h

    def _request(self, method: str, path: str, **kwargs) -> Optional[Any]:
        if not self.base_url:
            return None
        url = f"{self.base_url}/api{path}"
        try:
            with httpx.Client(timeout=self.timeout, verify=False) as client:
                resp = client.request(method, url, headers=self._headers(), **kwargs)
                resp.raise_for_status()
                if resp.status_code == 204 or not resp.content:
                    return {}
                return resp.json()
        except httpx.HTTPStatusError as e:
            logger.error(f"Semaphore HTTP {e.response.status_code} {url}: {e.response.text}")
            return None
        except httpx.HTTPError as e:
            logger.error(f"Semaphore request error {url}: {e}")
            return None

    # ── Projets ───────────────────────────────────────────────────────────────

    def get_projects(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        data = self._request("GET", "/projects")
        if data is None:
            return None
        results = [self._map_project(p) for p in (data if isinstance(data, list) else [])]
        return {"count": len(results), "results": results}

    def get_project(self, project_id: int) -> Optional[Dict[str, Any]]:
        data = self._request("GET", f"/project/{project_id}")
        return self._map_project(data) if data else None

    def _map_project(self, p: Dict) -> Dict:
        return {
            "id": p.get("id"),
            "name": p.get("name", ""),
            "description": p.get("description", ""),
            "scm_type": "git",
        }

    # ── Jobs (Tasks Semaphore) ─────────────────────────────────────────────────

    def get_jobs(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        pid = self.project_id
        data = self._request("GET", f"/project/{pid}/tasks?sort=id&order=desc&limit={page_size}")
        if data is None:
            return None
        templates_by_id = self._get_templates_by_id()
        results = [self._map_task(t, templates_by_id) for t in (data if isinstance(data, list) else [])]
        return {"count": len(results), "results": results}

    def get_job(self, job_id: int) -> Optional[Dict[str, Any]]:
        pid = self.project_id
        data = self._request("GET", f"/project/{pid}/tasks/{job_id}")
        return self._map_task(data, self._get_templates_by_id()) if data else None

    def get_job_stdout(self, job_id: int) -> Optional[str]:
        pid = self.project_id
        data = self._request("GET", f"/project/{pid}/tasks/{job_id}/output")
        if data is None:
            return None
        if isinstance(data, list):
            return "\n".join(entry.get("output", "") for entry in data)
        return str(data)

    def get_job_history(self, template_id: Optional[int] = None,
                        page: int = 1, page_size: int = 50) -> Optional[Dict[str, Any]]:
        return self.get_jobs(page, page_size)

    def _compute_elapsed(self, started: Optional[str], finished: Optional[str]) -> float:
        """Calcule la durée en secondes à partir de started/finished si Semaphore ne fournit pas duration."""
        if not started or not finished:
            return 0
        try:
            start_dt = datetime.fromisoformat(started.replace("Z", "+00:00"))
            end_dt = datetime.fromisoformat(finished.replace("Z", "+00:00"))
            return max((end_dt - start_dt).total_seconds(), 0)
        except (ValueError, AttributeError):
            return 0

    def _get_templates_by_id(self) -> Dict[int, Dict]:
        """Récupère les templates du projet indexés par id (pour enrichir les tasks)."""
        pid = self.project_id
        data = self._request("GET", f"/project/{pid}/templates")
        if not isinstance(data, list):
            return {}
        return {t.get("id"): t for t in data if t.get("id") is not None}

    def _get_inventories_by_id(self) -> Dict[int, Dict]:
        """Récupère les inventaires du projet indexés par id (pour enrichir templates/tasks)."""
        pid = self.project_id
        data = self._request("GET", f"/project/{pid}/inventory")
        if not isinstance(data, list):
            return {}
        return {i.get("id"): i for i in data if i.get("id") is not None}

    def _map_task(self, t: Dict, templates_by_id: Optional[Dict[int, Dict]] = None) -> Dict:
        status = STATUS_MAP.get(t.get("status", ""), t.get("status", "unknown"))
        templates_by_id = templates_by_id if templates_by_id is not None else self._get_templates_by_id()
        template = templates_by_id.get(t.get("template_id"), {}) or {}
        template_name = template.get("alias") or template.get("name") or ""

        inventory_name = None
        inventory_id = template.get("inventory_id")
        if inventory_id is not None:
            inventory = self._get_inventories_by_id().get(inventory_id)
            inventory_name = inventory.get("name") if inventory else None

        project_name_data = self.get_project(self.project_id)
        project_name = project_name_data.get("name") if project_name_data else None

        elapsed = t.get("duration") or self._compute_elapsed(t.get("created"), t.get("end"))

        return {
            "id": t.get("id"),
            "name": template_name or f"Task #{t.get('id')}",
            "status": status,
            "started": t.get("created"),
            "finished": t.get("end"),
            "elapsed": elapsed,
            "job_template": t.get("template_id"),
            "job_template_name": template_name,
            "inventory": inventory_id,
            "inventory_name": inventory_name,
            "project": self.project_id,
            "project_name": project_name,
            "summary_fields": {
                "job_template": {"name": template_name} if template_name else None,
                "inventory": {"name": inventory_name} if inventory_name else None,
                "project": {"name": project_name} if project_name else None,
            },
        }

    # ── Templates ─────────────────────────────────────────────────────────────

    def get_job_templates(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        pid = self.project_id
        data = self._request("GET", f"/project/{pid}/templates")
        if data is None:
            return None
        results = [self._map_template(t) for t in (data if isinstance(data, list) else [])]
        return {"count": len(results), "results": results}

    def get_job_template(self, template_id: int) -> Optional[Dict[str, Any]]:
        pid = self.project_id
        data = self._request("GET", f"/project/{pid}/templates/{template_id}")
        return self._map_template(data) if data else None

    def launch_job_template(self, template_id: int,
                            extra_vars: Optional[Dict[str, Any]] = None) -> Optional[Dict[str, Any]]:
        pid = self.project_id
        payload: Dict[str, Any] = {"template_id": template_id}
        if extra_vars:
            import json
            payload["environment"] = json.dumps(extra_vars)
        data = self._request("POST", f"/project/{pid}/tasks", json=payload)
        return self._map_task(data) if data else None

    def _map_template(self, t: Dict) -> Dict:
        inventory_id = t.get("inventory_id")
        inventory_name = None
        if inventory_id is not None:
            inventory = self._get_inventories_by_id().get(inventory_id)
            inventory_name = inventory.get("name") if inventory else None

        project_id = t.get("project_id") or self.project_id
        project_data = self.get_project(project_id)
        project_name = project_data.get("name") if project_data else None

        return {
            "id": t.get("id"),
            "name": t.get("alias") or t.get("name", ""),
            "description": t.get("description", ""),
            "playbook": t.get("playbook", ""),
            "inventory": inventory_id,
            "inventory_name": inventory_name,
            "project": project_id,
            "project_name": project_name,
            "summary_fields": {
                "inventory": {"name": inventory_name} if inventory_name else None,
                "project": {"name": project_name} if project_name else None,
            },
        }

    # ── Inventaires ───────────────────────────────────────────────────────────

    def get_inventories(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        pid = self.project_id
        data = self._request("GET", f"/project/{pid}/inventory")
        if data is None:
            return None
        results = [self._map_inventory(i) for i in (data if isinstance(data, list) else [])]
        return {"count": len(results), "results": results}

    def get_inventory(self, inventory_id: int) -> Optional[Dict[str, Any]]:
        pid = self.project_id
        data = self._request("GET", f"/project/{pid}/inventory/{inventory_id}")
        return self._map_inventory(data) if data else None

    def get_inventory_hosts(self, inventory_id: int,
                            page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        inventory = self.get_inventory(inventory_id)
        if not inventory:
            return {"count": 0, "results": []}
        content = inventory.get("_raw_inventory", "")
        hosts = [{"name": line.strip()} for line in content.splitlines()
                 if line.strip() and not line.startswith("[") and not line.startswith("#")]
        return {"count": len(hosts), "results": hosts}

    def _map_inventory(self, i: Dict) -> Dict:
        return {
            "id": i.get("id"),
            "name": i.get("name", ""),
            "type": i.get("type", "file"),
            "_raw_inventory": i.get("inventory", ""),
            "host_filter": None,
        }

    # ── Stubs compatibilité AWX ───────────────────────────────────────────────

    def get_credentials(self, **_) -> Dict:
        return {"count": 0, "results": []}

    def get_organizations(self, **_) -> Dict:
        return {"count": 1, "results": [{"id": 1, "name": "Default"}]}

    def get_playbooks(self, project_id: int) -> List[Dict]:
        return []
