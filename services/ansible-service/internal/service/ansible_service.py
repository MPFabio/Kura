"""Service métier pour Ansible Tower."""
import logging
import httpx
from typing import Optional, List, Dict, Any
from datetime import datetime

from internal.config.config import Config
from internal.client.tower_client import AnsibleTowerClient
from internal.cache.redis import RedisClient
from internal.configstore import ConfigstoreClient
from internal.models.models import (
    JobSummary,
    JobDetail,
    InventorySummary,
    InventoryDetail,
    Host,
    JobTemplateSummary,
    JobTemplateDetail,
    ProjectSummary,
    PlaybookInfo,
    JobHistoryResponse,
)

logger = logging.getLogger(__name__)


class AnsibleService:
    """Service métier pour gérer les interactions avec Ansible Tower."""

    def __init__(self, tower_client: AnsibleTowerClient, cache: Optional[RedisClient], config: Config):
        """Initialise le service Ansible."""
        self.tower_client = tower_client
        self.cache = cache
        self.config = config
        self.cfg_store = ConfigstoreClient(config.auth_service_url, "ansible")

    def _get_cache_key(self, prefix: str, *args) -> str:
        """Génère une clé de cache."""
        return f"ansible:{prefix}:{':'.join(str(a) for a in args)}"

    # ── Configuration Semaphore (persistée dans Postgres via configstore) ────────

    def get_config(self) -> Dict[str, Any]:
        """Retourne la configuration Semaphore active."""
        semaphore_url = self.cfg_store.get_or_fallback("semaphore_url", self.config.semaphore_url or "")
        project_id_str = self.cfg_store.get_or_fallback("semaphore_project_id", str(self.config.semaphore_project_id))
        token = self.cfg_store.get("semaphore_token") or self.config.semaphore_api_token or ""
        has_token = bool(token)
        try:
            project_id = int(project_id_str)
        except (ValueError, TypeError):
            project_id = 1
        return {
            "semaphore_url": semaphore_url,
            "semaphore_project_id": project_id,
            "has_token": has_token,
            "configured": bool(semaphore_url and has_token),
        }

    def set_config(self, semaphore_url: str = "", token: str = "", project_id: int = 1) -> Dict[str, Any]:
        """Met à jour la configuration Semaphore et réinitialise le client."""
        from internal.client.semaphore_client import SemaphoreClient

        kv: Dict[str, str] = {"semaphore_project_id": str(project_id)}
        if semaphore_url:
            kv["semaphore_url"] = semaphore_url
        if token:
            kv["semaphore_token"] = token
        self.cfg_store.set_many(kv)

        # Mise à jour de la config en mémoire
        if semaphore_url:
            self.config.semaphore_url = semaphore_url
        if token:
            self.config.semaphore_api_token = token
        self.config.semaphore_project_id = project_id

        # Réinitialisation du client à chaud
        self.tower_client = SemaphoreClient(self.config)
        logger.info(f"Client Semaphore réinitialisé → {semaphore_url}, projet {project_id}")

        return self.get_config()

    def get_jobs(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des jobs avec cache."""
        from internal.metrics.prometheus import cache_hits_total, cache_misses_total

        cache_key = self._get_cache_key("jobs", page, page_size)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                logger.debug(f"Jobs récupérés depuis le cache: page={page}")
                cache_hits_total.labels(cache_type="jobs").inc()
                return cached
            cache_misses_total.labels(cache_type="jobs").inc()

        result = self.tower_client.get_jobs(page, page_size)
        if result and self.cache:
            self.cache.set_json(cache_key, result)
        return result

    def get_job(self, job_id: int, include_stdout: bool = False) -> Optional[JobDetail]:
        """Récupère les détails d'un job spécifique."""
        cache_key = self._get_cache_key("job", job_id, include_stdout)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                logger.debug(f"Job {job_id} récupéré depuis le cache")
                return JobDetail(**cached)

        job_data = self.tower_client.get_job(job_id)
        if not job_data:
            return None

        # Récupérer la sortie standard si demandée
        stdout = None
        if include_stdout:
            stdout = self.tower_client.get_job_stdout(job_id)

        # Construire le modèle JobDetail
        job_detail = JobDetail(
            id=job_data.get("id"),
            name=job_data.get("name", ""),
            status=job_data.get("status", "unknown"),
            started=self._parse_datetime(job_data.get("started")),
            finished=self._parse_datetime(job_data.get("finished")),
            elapsed=job_data.get("elapsed"),
            job_template=job_data.get("job_template"),
            job_template_name=self._extract_name(job_data.get("summary_fields", {}).get("job_template")),
            inventory=job_data.get("inventory"),
            inventory_name=self._extract_name(job_data.get("summary_fields", {}).get("inventory")),
            project=job_data.get("project"),
            project_name=self._extract_name(job_data.get("summary_fields", {}).get("project")),
            playbook=job_data.get("playbook"),
            limit=job_data.get("limit"),
            verbosity=job_data.get("verbosity"),
            extra_vars=self._parse_variables(job_data.get("extra_vars"))
            if isinstance(job_data.get("extra_vars"), str)
            else job_data.get("extra_vars"),
            created_by=job_data.get("created_by"),
            created_by_username=self._extract_name(job_data.get("summary_fields", {}).get("created_by")),
            stdout=stdout,
        )

        # Mettre en cache (TTL plus court pour les détails) - mode='json' pour datetime
        if self.cache:
            self.cache.set_json(cache_key, job_detail.model_dump(mode='json'), ttl=60)
        return job_detail

    def get_inventories(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des inventaires avec cache."""
        cache_key = self._get_cache_key("inventories", page, page_size)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                logger.debug(f"Inventaires récupérés depuis le cache: page={page}")
                return cached

        result = self.tower_client.get_inventories(page, page_size)
        if result and self.cache:
            self.cache.set_json(cache_key, result)
        return result

    def get_inventory(self, inventory_id: int) -> Optional[InventoryDetail]:
        """Récupère les détails d'un inventaire spécifique."""
        cache_key = self._get_cache_key("inventory", inventory_id)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                logger.debug(f"Inventaire {inventory_id} récupéré depuis le cache")
                return InventoryDetail(**cached)

        inventory_data = self.tower_client.get_inventory(inventory_id)
        if not inventory_data:
            return None

        inventory_detail = InventoryDetail(
            id=inventory_data.get("id"),
            name=inventory_data.get("name", ""),
            description=inventory_data.get("description"),
            organization=inventory_data.get("organization"),
            organization_name=self._extract_name(inventory_data.get("summary_fields", {}).get("organization")),
            kind=inventory_data.get("kind"),
            host_count=inventory_data.get("host_count", 0),
            created=self._parse_datetime(inventory_data.get("created")),
            modified=self._parse_datetime(inventory_data.get("modified")),
            variables=self._parse_variables(inventory_data.get("variables")),
        )

        if self.cache:
            self.cache.set_json(cache_key, inventory_detail.model_dump(), ttl=300)
        return inventory_detail

    def get_inventory_hosts(self, inventory_id: int, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère les hôtes d'un inventaire."""
        cache_key = self._get_cache_key("inventory_hosts", inventory_id, page, page_size)
        cached = self.cache.get_json(cache_key)
        if cached:
            return cached

        result = self.tower_client.get_inventory_hosts(inventory_id, page, page_size)
        if result:
            self.cache.set_json(cache_key, result, ttl=180)
        return result

    def get_job_templates(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des templates de jobs avec cache."""
        cache_key = self._get_cache_key("job_templates", page, page_size)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                logger.debug(f"Templates de jobs récupérés depuis le cache: page={page}")
                return cached

        result = self.tower_client.get_job_templates(page, page_size)
        if result and self.cache:
            self.cache.set_json(cache_key, result)
        return result

    def get_job_template(self, template_id: int) -> Optional[JobTemplateDetail]:
        """Récupère les détails d'un template de job spécifique."""
        cache_key = self._get_cache_key("job_template", template_id)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                return JobTemplateDetail(**cached)

        template_data = self.tower_client.get_job_template(template_id)
        if not template_data:
            return None

        template_detail = JobTemplateDetail(
            id=template_data.get("id"),
            name=template_data.get("name", ""),
            description=template_data.get("description"),
            job_type=template_data.get("job_type"),
            inventory=template_data.get("inventory"),
            inventory_name=self._extract_name(template_data.get("summary_fields", {}).get("inventory")),
            project=template_data.get("project"),
            project_name=self._extract_name(template_data.get("summary_fields", {}).get("project")),
            playbook=template_data.get("playbook"),
            limit=template_data.get("limit"),
            verbosity=template_data.get("verbosity"),
            extra_vars=self._parse_variables(template_data.get("extra_vars")),
            ask_variables_on_launch=template_data.get("ask_variables_on_launch", False),
            ask_limit_on_launch=template_data.get("ask_limit_on_launch", False),
            ask_tags_on_launch=template_data.get("ask_tags_on_launch", False),
            ask_skip_tags_on_launch=template_data.get("ask_skip_tags_on_launch", False),
            ask_job_type_on_launch=template_data.get("ask_job_type_on_launch", False),
            ask_verbosity_on_launch=template_data.get("ask_verbosity_on_launch", False),
            ask_inventory_on_launch=template_data.get("ask_inventory_on_launch", False),
            ask_credential_on_launch=template_data.get("ask_credential_on_launch", False),
            created=self._parse_datetime(template_data.get("created")),
            modified=self._parse_datetime(template_data.get("modified")),
        )

        if self.cache:
            self.cache.set_json(cache_key, template_detail.model_dump(), ttl=300)
        return template_detail

    def launch_job_template(self, template_id: int, extra_vars: Optional[Dict[str, Any]] = None) -> Optional[JobDetail]:
        """Lance un job depuis un template."""
        extra_vars = dict(extra_vars) if extra_vars else {}

        # Injecter les informations nécessaires aux playbooks pour accéder au cluster k8s actif
        if self.config.internal_api_secret:
            try:
                with httpx.Client(timeout=5.0) as client:
                    resp = client.get(f"{self.config.k8s_service_url}/api/v1/k8s/clusters/active")
                    if resp.status_code == 200:
                        cluster = resp.json()
                        cluster_id = cluster.get("id")
                        if cluster_id:
                            extra_vars["cluster_id"] = cluster_id
                            extra_vars["k8s_service_url"] = self.config.k8s_service_url
                            extra_vars["internal_api_token"] = self.config.internal_api_secret
            except Exception as e:
                logger.warning(f"Impossible de récupérer le cluster actif pour injection dans le job: {e}")

        result = self.tower_client.launch_job_template(template_id, extra_vars)
        if not result:
            return None

        # Invalider le cache des jobs
        if self.cache:
            for key in self.cache.keys("ansible:jobs:*"):
                self.cache.delete(key)

        # Retourner les détails du job lancé
        job_id = result.get("id")
        if job_id:
            return self.get_job(job_id)

        return None

    def get_projects(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des projets."""
        cache_key = self._get_cache_key("projects", page, page_size)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                return cached

        result = self.tower_client.get_projects(page, page_size)
        if result and self.cache:
            self.cache.set_json(cache_key, result)
        return result

    def get_playbooks(self, project_id: int) -> List[PlaybookInfo]:
        """Récupère la liste des playbooks d'un projet."""
        cache_key = self._get_cache_key("playbooks", project_id)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                return [PlaybookInfo(**p) for p in cached]

        playbooks_data = self.tower_client.get_playbooks(project_id)
        if not playbooks_data:
            return []

        project = self.tower_client.get_project(project_id)
        playbooks = []
        for pb_data in playbooks_data:
            playbook = PlaybookInfo(
                name=pb_data.get("name", "unknown"),
                project_id=project_id,
                project_name=project.get("name") if project else None,
                scm_type=pb_data.get("scm_type"),
                scm_url=pb_data.get("scm_url"),
            )
            playbooks.append(playbook)

        if self.cache:
            self.cache.set_json(cache_key, [p.model_dump() for p in playbooks], ttl=600)
        return playbooks

    def get_job_history(self, template_id: Optional[int] = None, page: int = 1, page_size: int = 50) -> Optional[JobHistoryResponse]:
        """Récupère l'historique des jobs."""
        cache_key = self._get_cache_key("job_history", template_id or "all", page, page_size)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                return JobHistoryResponse(**cached)

        result = self.tower_client.get_job_history(template_id, page, page_size)
        if not result:
            return None

        jobs = []
        for job_data in result.get("results", []):
            job = JobSummary(
                id=job_data.get("id"),
                name=job_data.get("name", ""),
                status=job_data.get("status", "unknown"),
                started=self._parse_datetime(job_data.get("started")),
                finished=self._parse_datetime(job_data.get("finished")),
                elapsed=job_data.get("elapsed"),
                job_template=job_data.get("job_template"),
                job_template_name=self._extract_name(job_data.get("summary_fields", {}).get("job_template")),
                inventory=job_data.get("inventory"),
                inventory_name=self._extract_name(job_data.get("summary_fields", {}).get("inventory")),
                created_by=job_data.get("created_by"),
                created_by_username=self._extract_name(job_data.get("summary_fields", {}).get("created_by")),
            )
            jobs.append(job)

        history_response = JobHistoryResponse(
            count=result.get("count", 0),
            results=jobs,
            next=result.get("next"),
            previous=result.get("previous"),
        )

        if self.cache:
            self.cache.set_json(cache_key, history_response.model_dump(), ttl=120)
        return history_response

    def _parse_datetime(self, value: Optional[str]) -> Optional[datetime]:
        """Parse une chaîne datetime depuis l'API Ansible Tower."""
        if not value:
            return None
        try:
            # Format ISO 8601 avec timezone
            return datetime.fromisoformat(value.replace("Z", "+00:00"))
        except (ValueError, AttributeError):
            return None

    def _extract_name(self, obj: Optional[Dict[str, Any]]) -> Optional[str]:
        """Extrait le nom d'un objet depuis summary_fields."""
        if not obj:
            return None
        if isinstance(obj, dict):
            return obj.get("name")
        return None

    def _parse_variables(self, value: Optional[str]) -> Optional[Dict[str, Any]]:
        """Parse les variables JSON depuis Ansible Tower."""
        if not value:
            return None
        try:
            import json
            return json.loads(value)
        except (json.JSONDecodeError, TypeError):
            return None

    # Credentials
    def get_credentials(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des credentials avec cache."""
        cache_key = self._get_cache_key("credentials", page, page_size)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                return cached

        result = self.tower_client.get_credentials(page, page_size)
        if result and self.cache:
            self.cache.set_json(cache_key, result)
        return result

    def get_credential(self, credential_id: int):
        """Récupère les détails d'un credential."""
        from internal.models.models import CredentialDetail

        cache_key = self._get_cache_key("credential", credential_id)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                return CredentialDetail(**cached)

        credential_data = self.tower_client.get_credential(credential_id)
        if not credential_data:
            return None

        credential_detail = CredentialDetail(
            id=credential_data.get("id"),
            name=credential_data.get("name", ""),
            description=credential_data.get("description"),
            credential_type=credential_data.get("credential_type"),
            credential_type_name=self._extract_name(credential_data.get("summary_fields", {}).get("credential_type")),
            organization=credential_data.get("organization"),
            organization_name=self._extract_name(credential_data.get("summary_fields", {}).get("organization")),
            inputs=credential_data.get("inputs"),  # Note: peut être masqué par sécurité
            created=self._parse_datetime(credential_data.get("created")),
            modified=self._parse_datetime(credential_data.get("modified")),
        )

        if self.cache:
            self.cache.set_json(cache_key, credential_detail.model_dump(), ttl=300)
        return credential_detail

    def create_credential(self, credential_data: Dict[str, Any]):
        """Crée un nouveau credential."""
        result = self.tower_client.create_credential(credential_data)
        if result and self.cache:
            # Invalider le cache des credentials
            for key in self.cache.keys("ansible:credentials:*"):
                self.cache.delete(key)
        return result

    def update_credential(self, credential_id: int, credential_data: Dict[str, Any]):
        """Met à jour un credential."""
        result = self.tower_client.update_credential(credential_id, credential_data)
        if result and self.cache:
            # Invalider le cache
            self.cache.delete(self._get_cache_key("credential", credential_id))
            for key in self.cache.keys("ansible:credentials:*"):
                self.cache.delete(key)
        return result

    def delete_credential(self, credential_id: int) -> bool:
        """Supprime un credential."""
        result = self.tower_client.delete_credential(credential_id)
        if result and self.cache:
            # Invalider le cache
            self.cache.delete(self._get_cache_key("credential", credential_id))
            for key in self.cache.keys("ansible:credentials:*"):
                self.cache.delete(key)
        return result

    # Organizations
    def get_organizations(self, page: int = 1, page_size: int = 20) -> Optional[Dict[str, Any]]:
        """Récupère la liste des organisations avec cache."""
        cache_key = self._get_cache_key("organizations", page, page_size)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                return cached

        result = self.tower_client.get_organizations(page, page_size)
        if result and self.cache:
            self.cache.set_json(cache_key, result)
        return result

    def get_organization(self, organization_id: int):
        """Récupère les détails d'une organisation."""
        from internal.models.models import OrganizationDetail

        cache_key = self._get_cache_key("organization", organization_id)
        if self.cache:
            cached = self.cache.get_json(cache_key)
            if cached:
                return OrganizationDetail(**cached)

        org_data = self.tower_client.get_organization(organization_id)
        if not org_data:
            return None

        org_detail = OrganizationDetail(
            id=org_data.get("id"),
            name=org_data.get("name", ""),
            description=org_data.get("description"),
            max_hosts=org_data.get("max_hosts"),
            custom_virtualenv=org_data.get("custom_virtualenv"),
            default_environment=org_data.get("default_environment"),
            created=self._parse_datetime(org_data.get("created")),
            modified=self._parse_datetime(org_data.get("modified")),
        )

        if self.cache:
            self.cache.set_json(cache_key, org_detail.model_dump(), ttl=300)
        return org_detail

    def create_organization(self, org_data: Dict[str, Any]):
        """Crée une nouvelle organisation."""
        result = self.tower_client.create_organization(org_data)
        if result and self.cache:
            # Invalider le cache
            for key in self.cache.keys("ansible:organizations:*"):
                self.cache.delete(key)
        return result

    def update_organization(self, organization_id: int, org_data: Dict[str, Any]):
        """Met à jour une organisation."""
        result = self.tower_client.update_organization(organization_id, org_data)
        if result and self.cache:
            # Invalider le cache
            self.cache.delete(self._get_cache_key("organization", organization_id))
            for key in self.cache.keys("ansible:organizations:*"):
                self.cache.delete(key)
        return result

    def delete_organization(self, organization_id: int) -> bool:
        """Supprime une organisation."""
        result = self.tower_client.delete_organization(organization_id)
        if result and self.cache:
            # Invalider le cache
            self.cache.delete(self._get_cache_key("organization", organization_id))
            for key in self.cache.keys("ansible:organizations:*"):
                self.cache.delete(key)
        return result

    # Playbook analysis
    def analyze_playbook(self, playbook_content: str):
        """Analyse un playbook en profondeur."""
        from internal.parser.playbook_parser import PlaybookParser
        return PlaybookParser.analyze_playbook(playbook_content)
