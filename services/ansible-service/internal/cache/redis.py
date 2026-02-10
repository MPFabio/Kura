"""Client Redis pour le cache."""
import json
import logging
from typing import Optional, Any
import redis
from redis.exceptions import RedisError

from internal.config.config import Config

logger = logging.getLogger(__name__)


class RedisClient:
    """Client Redis pour le cache."""

    def __init__(self, config: Config):
        """Initialise le client Redis."""
        self.config = config
        self.client: Optional[redis.Redis] = None
        self._connect()

    def _connect(self):
        """Établit la connexion à Redis."""
        try:
            self.client = redis.Redis(
                host=self.config.redis_host,
                port=int(self.config.redis_port),
                password=self.config.redis_password,
                db=self.config.redis_db,
                decode_responses=True,
                socket_connect_timeout=2,
            )
            # Tester la connexion
            self.client.ping()
            logger.info("Connexion Redis établie avec succès")
        except RedisError as e:
            logger.error(f"Erreur lors de la connexion à Redis: {e}")
            raise

    def get(self, key: str) -> Optional[str]:
        """Récupère une valeur depuis Redis."""
        try:
            if not self.client:
                return None
            return self.client.get(key)
        except RedisError as e:
            logger.warning(f"Erreur lors de la récupération de la clé {key}: {e}")
            return None

    def set(self, key: str, value: str, ttl: Optional[int] = None) -> bool:
        """Stocke une valeur dans Redis avec un TTL optionnel."""
        try:
            if not self.client:
                return False
            ttl = ttl or self.config.cache_ttl
            return self.client.set(key, value, ex=ttl)
        except RedisError as e:
            logger.warning(f"Erreur lors du stockage de la clé {key}: {e}")
            return False

    def get_json(self, key: str) -> Optional[Any]:
        """Récupère une valeur JSON depuis Redis."""
        value = self.get(key)
        if value:
            try:
                return json.loads(value)
            except json.JSONDecodeError:
                return None
        return None

    def set_json(self, key: str, value: Any, ttl: Optional[int] = None) -> bool:
        """Stocke une valeur JSON dans Redis."""
        try:
            json_value = json.dumps(value)
            return self.set(key, json_value, ttl)
        except (TypeError, ValueError) as e:
            logger.warning(f"Erreur lors de la sérialisation JSON pour {key}: {e}")
            return False

    def delete(self, key: str) -> bool:
        """Supprime une clé de Redis."""
        try:
            if not self.client:
                return False
            return bool(self.client.delete(key))
        except RedisError as e:
            logger.warning(f"Erreur lors de la suppression de la clé {key}: {e}")
            return False

    def keys(self, pattern: str) -> list[str]:
        """Retourne toutes les clés correspondant au pattern."""
        try:
            if not self.client:
                return []
            return list(self.client.keys(pattern))
        except RedisError as e:
            logger.warning(f"Erreur lors de la recherche de clés avec le pattern {pattern}: {e}")
            return []

    def close(self):
        """Ferme la connexion Redis."""
        if self.client:
            self.client.close()
            logger.info("Connexion Redis fermée")
