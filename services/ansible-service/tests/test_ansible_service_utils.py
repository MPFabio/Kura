"""Tests unitaires des helpers d'AnsibleService (parse, extract, etc.)."""

from datetime import datetime, timezone

from internal.config.config import Config
from internal.service.ansible_service import AnsibleService


class _DummyTowerClient:
    """Client Tower minimal pour les tests qui ne touchent pas l'API."""

    pass


def _make_service() -> AnsibleService:
    return AnsibleService(
        tower_client=_DummyTowerClient(),
        cache=None,
        config=Config(),
    )


def test_parse_datetime_with_valid_iso_string():
    service = _make_service()

    value = "2024-01-01T12:00:00Z"
    dt = service._parse_datetime(value)

    assert isinstance(dt, datetime)
    # La valeur "Z" doit être interprétée comme UTC
    assert dt.tzinfo is not None
    assert dt.tzinfo == timezone.utc


def test_parse_datetime_with_invalid_value_returns_none():
    service = _make_service()

    assert service._parse_datetime("not-a-date") is None
    assert service._parse_datetime(None) is None


def test_extract_name_from_dict_and_none():
    service = _make_service()

    assert service._extract_name({"name": "my-object"}) == "my-object"
    assert service._extract_name({"other": "value"}) is None
    assert service._extract_name(None) is None


def test_parse_variables_parses_valid_json_and_handles_errors():
    service = _make_service()

    parsed = service._parse_variables('{"key": "value", "n": 1}')
    assert isinstance(parsed, dict)
    assert parsed["key"] == "value"
    assert parsed["n"] == 1

    # Valeurs invalides ou vides -> None
    assert service._parse_variables("not-json") is None
    assert service._parse_variables(None) is None

