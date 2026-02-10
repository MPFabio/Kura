"""Modèles de données pour le service Ansible."""
from datetime import datetime
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field


class JobSummary(BaseModel):
    """Résumé d'un job Ansible."""
    id: int
    name: str
    status: str
    started: Optional[datetime] = None
    finished: Optional[datetime] = None
    elapsed: Optional[float] = None
    job_template: Optional[int] = None
    job_template_name: Optional[str] = None
    inventory: Optional[int] = None
    inventory_name: Optional[str] = None
    created_by: Optional[int] = None
    created_by_username: Optional[str] = None


class JobDetail(BaseModel):
    """Détails complets d'un job Ansible."""
    id: int
    name: str
    status: str
    started: Optional[datetime] = None
    finished: Optional[datetime] = None
    elapsed: Optional[float] = None
    job_template: Optional[int] = None
    job_template_name: Optional[str] = None
    inventory: Optional[int] = None
    inventory_name: Optional[str] = None
    project: Optional[int] = None
    project_name: Optional[str] = None
    playbook: Optional[str] = None
    limit: Optional[str] = None
    verbosity: Optional[int] = None
    extra_vars: Optional[Dict[str, Any]] = None
    created_by: Optional[int] = None
    created_by_username: Optional[str] = None
    stdout: Optional[str] = None


class InventorySummary(BaseModel):
    """Résumé d'un inventaire Ansible."""
    id: int
    name: str
    description: Optional[str] = None
    organization: Optional[int] = None
    organization_name: Optional[str] = None
    kind: Optional[str] = None  # "", "smart", etc.
    host_count: Optional[int] = None
    created: Optional[datetime] = None
    modified: Optional[datetime] = None


class InventoryDetail(BaseModel):
    """Détails complets d'un inventaire Ansible."""
    id: int
    name: str
    description: Optional[str] = None
    organization: Optional[int] = None
    organization_name: Optional[str] = None
    kind: Optional[str] = None
    host_count: Optional[int] = None
    created: Optional[datetime] = None
    modified: Optional[datetime] = None
    variables: Optional[Dict[str, Any]] = None


class Host(BaseModel):
    """Hôte Ansible."""
    id: int
    name: str
    description: Optional[str] = None
    inventory: int
    enabled: bool = True
    variables: Optional[Dict[str, Any]] = None
    created: Optional[datetime] = None
    modified: Optional[datetime] = None


class JobTemplateSummary(BaseModel):
    """Résumé d'un template de job Ansible."""
    id: int
    name: str
    description: Optional[str] = None
    job_type: Optional[str] = None  # "run", "check", etc.
    inventory: Optional[int] = None
    inventory_name: Optional[str] = None
    project: Optional[int] = None
    project_name: Optional[str] = None
    playbook: Optional[str] = None
    created: Optional[datetime] = None
    modified: Optional[datetime] = None


class JobTemplateDetail(BaseModel):
    """Détails complets d'un template de job Ansible."""
    id: int
    name: str
    description: Optional[str] = None
    job_type: Optional[str] = None
    inventory: Optional[int] = None
    inventory_name: Optional[str] = None
    project: Optional[int] = None
    project_name: Optional[str] = None
    playbook: Optional[str] = None
    limit: Optional[str] = None
    verbosity: Optional[int] = None
    extra_vars: Optional[Dict[str, Any]] = None
    ask_variables_on_launch: bool = False
    ask_limit_on_launch: bool = False
    ask_tags_on_launch: bool = False
    ask_skip_tags_on_launch: bool = False
    ask_job_type_on_launch: bool = False
    ask_verbosity_on_launch: bool = False
    ask_inventory_on_launch: bool = False
    ask_credential_on_launch: bool = False
    created: Optional[datetime] = None
    modified: Optional[datetime] = None


class ProjectSummary(BaseModel):
    """Résumé d'un projet Ansible."""
    id: int
    name: str
    description: Optional[str] = None
    scm_type: Optional[str] = None  # "git", "svn", "manual", etc.
    scm_url: Optional[str] = None
    scm_branch: Optional[str] = None
    status: Optional[str] = None
    created: Optional[datetime] = None
    modified: Optional[datetime] = None


class PlaybookInfo(BaseModel):
    """Informations sur un playbook."""
    name: str
    project_id: int
    project_name: Optional[str] = None
    scm_type: Optional[str] = None
    scm_url: Optional[str] = None
    path: Optional[str] = None


class JobHistoryResponse(BaseModel):
    """Réponse pour l'historique des jobs."""
    count: int
    results: List[JobSummary]
    next: Optional[str] = None
    previous: Optional[str] = None


class PaginatedResponse(BaseModel):
    """Réponse paginée générique."""
    count: int
    results: List[Any]
    next: Optional[str] = None
    previous: Optional[str] = None


class CredentialSummary(BaseModel):
    """Résumé d'un credential Ansible."""
    id: int
    name: str
    description: Optional[str] = None
    credential_type: Optional[int] = None
    credential_type_name: Optional[str] = None
    organization: Optional[int] = None
    organization_name: Optional[str] = None
    created: Optional[datetime] = None
    modified: Optional[datetime] = None


class CredentialDetail(BaseModel):
    """Détails complets d'un credential Ansible."""
    id: int
    name: str
    description: Optional[str] = None
    credential_type: Optional[int] = None
    credential_type_name: Optional[str] = None
    organization: Optional[int] = None
    organization_name: Optional[str] = None
    inputs: Optional[Dict[str, Any]] = None  # Les inputs sont masqués par sécurité
    created: Optional[datetime] = None
    modified: Optional[datetime] = None


class CredentialCreate(BaseModel):
    """Modèle pour créer un credential."""
    name: str
    description: Optional[str] = None
    credential_type: int
    organization: Optional[int] = None
    inputs: Dict[str, Any] = Field(default_factory=dict)


class CredentialUpdate(BaseModel):
    """Modèle pour mettre à jour un credential."""
    name: Optional[str] = None
    description: Optional[str] = None
    credential_type: Optional[int] = None
    organization: Optional[int] = None
    inputs: Optional[Dict[str, Any]] = None


class OrganizationSummary(BaseModel):
    """Résumé d'une organisation Ansible."""
    id: int
    name: str
    description: Optional[str] = None
    created: Optional[datetime] = None
    modified: Optional[datetime] = None


class OrganizationDetail(BaseModel):
    """Détails complets d'une organisation Ansible."""
    id: int
    name: str
    description: Optional[str] = None
    max_hosts: Optional[int] = None
    custom_virtualenv: Optional[str] = None
    default_environment: Optional[int] = None
    created: Optional[datetime] = None
    modified: Optional[datetime] = None


class OrganizationCreate(BaseModel):
    """Modèle pour créer une organisation."""
    name: str
    description: Optional[str] = None
    max_hosts: Optional[int] = None
    custom_virtualenv: Optional[str] = None
    default_environment: Optional[int] = None


class OrganizationUpdate(BaseModel):
    """Modèle pour mettre à jour une organisation."""
    name: Optional[str] = None
    description: Optional[str] = None
    max_hosts: Optional[int] = None
    custom_virtualenv: Optional[str] = None
    default_environment: Optional[int] = None


class PlaybookAnalysisRequest(BaseModel):
    """Requête pour analyser un playbook."""
    playbook_content: str


class PlaybookAnalysis(BaseModel):
    """Analyse complète d'un playbook."""
    parsed: Dict[str, Any]
    statistics: Dict[str, Any]


class WebhookEvent(BaseModel):
    """Événement webhook d'Ansible Tower."""
    event: str
    job_id: Optional[int] = None
    job_template_id: Optional[int] = None
    status: Optional[str] = None
    timestamp: Optional[datetime] = None
    data: Optional[Dict[str, Any]] = None
