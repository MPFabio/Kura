"""Client HTTP pour lire et écrire les configurations dans Postgres via auth-service."""
import logging
import urllib.request
import urllib.error
import json
from typing import Optional

logger = logging.getLogger(__name__)


class ConfigstoreClient:
    """Client configstore — appelle auth-service /internal/config/:service."""

    def __init__(self, auth_service_url: str, service: str):
        self._base = auth_service_url.rstrip("/")
        self._service = service

    def get(self, key: str) -> Optional[str]:
        """Lit une clé depuis Postgres. Retourne None si absente ou erreur."""
        try:
            url = f"{self._base}/internal/config/{self._service}/{key}"
            with urllib.request.urlopen(url, timeout=2) as resp:
                data = json.loads(resp.read())
                return data.get("value") or None
        except Exception as e:
            logger.debug("configstore.get %s: %s", key, e)
            return None

    def get_all(self) -> dict:
        """Lit toutes les clés du service. Retourne {} en cas d'erreur."""
        try:
            url = f"{self._base}/internal/config/{self._service}"
            with urllib.request.urlopen(url, timeout=2) as resp:
                data = json.loads(resp.read())
                return data.get("config", {})
        except Exception as e:
            logger.debug("configstore.get_all: %s", e)
            return {}

    def set_many(self, kv: dict) -> bool:
        """Écrit plusieurs clés dans Postgres. Retourne True si succès."""
        try:
            url = f"{self._base}/internal/config/{self._service}"
            payload = json.dumps(kv).encode()
            req = urllib.request.Request(url, data=payload, method="POST",
                                         headers={"Content-Type": "application/json"})
            with urllib.request.urlopen(req, timeout=2):
                return True
        except Exception as e:
            logger.warning("configstore.set_many: %s", e)
            return False

    def get_or_fallback(self, key: str, fallback: str = "") -> str:
        """Lit la clé dans Postgres, retourne fallback si vide/absent."""
        val = self.get(key)
        return val if val else fallback
