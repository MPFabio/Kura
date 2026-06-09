#!/bin/sh
# Initialise les secrets Kura dans Vault au premier démarrage (mode dev).
# Appelé manuellement ou via un init-container si besoin.
set -e

VAULT_ADDR="${VAULT_ADDR:-http://vault:8200}"
VAULT_TOKEN="${VAULT_TOKEN:-${VAULT_DEV_ROOT_TOKEN}}"

export VAULT_ADDR VAULT_TOKEN

echo "→ Activation du moteur KV v2..."
vault secrets enable -path=secret kv-v2 2>/dev/null || true

echo "→ Création des secrets Kura..."

vault kv put secret/kura/postgres \
  host=postgres \
  port=5432 \
  db="${POSTGRES_DB:-kura}" \
  user="${POSTGRES_USER:-kura}" \
  password="${POSTGRES_PASSWORD}"

vault kv put secret/kura/redis \
  host=redis \
  port=6379 \
  password="${REDIS_PASSWORD}"

vault kv put secret/kura/jwt \
  secret="${JWT_SECRET}" \
  expiration="24h"

vault kv put secret/kura/terraform \
  encryption_key="${TERRAFORM_ENCRYPTION_KEY}"

vault kv put secret/kura/pipeline \
  webhook_secret="${WEBHOOK_SECRET}" \
  github_token="${GITHUB_TOKEN:-}"

vault kv put secret/kura/grafana \
  admin_user="${GF_SECURITY_ADMIN_USER:-admin}" \
  admin_password="${GF_SECURITY_ADMIN_PASSWORD}"

echo "✅ Secrets Kura initialisés dans Vault."
