"""Configuration du service Ansible."""
import os
from typing import Optional


class Config:
    """Configuration du service Ansible Tower."""

    def __init__(
        self,
        server_port: str = "8083",
        environment: str = "development",
        log_level: str = "info",
        redis_host: str = "localhost",
        redis_port: str = "6379",
        redis_password: Optional[str] = None,
        redis_db: int = 0,
        cache_ttl: int = 300,
        ansible_tower_url: Optional[str] = None,
        ansible_tower_username: Optional[str] = None,
        ansible_tower_password: Optional[str] = None,
        ansible_tower_verify_ssl: bool = True,
    ):
        """Initialise la configuration."""
        self.server_port = server_port
        self.environment = environment
        self.log_level = log_level
        self.redis_host = redis_host
        self.redis_port = redis_port
        self.redis_password = redis_password
        self.redis_db = redis_db
        self.cache_ttl = cache_ttl
        self.ansible_tower_url = ansible_tower_url
        self.ansible_tower_username = ansible_tower_username
        self.ansible_tower_password = ansible_tower_password
        self.ansible_tower_verify_ssl = ansible_tower_verify_ssl


def load_config() -> Config:
    """Charge la configuration depuis les variables d'environnement."""
    return Config(
        server_port=os.getenv("ANSIBLE_SERVICE_PORT", "8083"),
        environment=os.getenv("ENV", "development"),
        log_level=os.getenv("LOG_LEVEL", "info"),
        redis_host=os.getenv("REDIS_HOST", "localhost"),
        redis_port=os.getenv("REDIS_PORT", "6379"),
        redis_password=os.getenv("REDIS_PASSWORD"),
        redis_db=int(os.getenv("REDIS_DB", "0")),
        cache_ttl=int(os.getenv("ANSIBLE_CACHE_TTL", "300")),
        ansible_tower_url=os.getenv("ANSIBLE_TOWER_URL"),
        ansible_tower_username=os.getenv("ANSIBLE_TOWER_USERNAME"),
        ansible_tower_password=os.getenv("ANSIBLE_TOWER_PASSWORD"),
        ansible_tower_verify_ssl=os.getenv("ANSIBLE_TOWER_VERIFY_SSL", "true").lower() in ("true", "1", "yes"),
    )
