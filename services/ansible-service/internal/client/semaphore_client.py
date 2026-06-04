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
        results = [self._map_task(t) for t in (data if isinstance(data, list) else [])]
        return {"count": len(results), "results": results}

    def get_job(self, job_id: int) -> Optional[Dict[str, Any]]:
        pid = self.project_id
        data = self._request("GET", f"/project/{pid}/tasks/{job_id}")
        return self._map_task(data) if data else None

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

    def _map_task(self, t: Dict) -> Dict:
        status = STATUS_MAP.get(t.get("status", ""), t.get("status", "unknown"))
        template = t.get("template", {}) or {}
        return {
            "id": t.get("id"),
            "name": template.get("alias") or template.get("name") or f"Task #{t.get('id')}",
            "status": status,
            "started": t.get("created"),
            "finished": t.get("end"),
            "job_template": t.get("template_id"),
            "job_template_name": template.get("alias") or template.get("name", ""),
            "elapsed": t.get("duration", 0),
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
        return {
            "id": t.get("id"),
            "name": t.get("alias") or t.get("name", ""),
            "description": t.get("description", ""),
            "playbook": t.get("playbook", ""),
            "inventory": t.get("inventory_id"),
            "project": t.get("project_id") or self.project_id,
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
