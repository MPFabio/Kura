#!/bin/sh
# Initialise les secrets Kura dans OpenBao au premier démarrage (mode dev).
# Appelé manuellement ou via un init-container si besoin.
set -e

BAO_ADDR="${OPENBAO_ADDR:-http://openbao:8200}"
BAO_TOKEN="${OPENBAO_TOKEN:-${OPENBAO_DEV_ROOT_TOKEN}}"

export BAO_ADDR BAO_TOKEN

echo "→ Activation du moteur KV v2..."
bao secrets enable -path=secret kv-v2 2>/dev/null || true

echo "→ Création des secrets Kura..."

bao kv put secret/kura/postgres \
  host=postgres \
  port=5432 \
  db="${POSTGRES_DB:-kura}" \
  user="${POSTGRES_USER:-kura}" \
  password="${POSTGRES_PASSWORD}"

bao kv put secret/kura/valkey \
  host=valkey \
  port=6379 \
  password="${VALKEY_PASSWORD}"

bao kv put secret/kura/jwt \
  secret="${JWT_SECRET}" \
  expiration="24h"

bao kv put secret/kura/terraform \
  encryption_key="${TERRAFORM_ENCRYPTION_KEY}"

bao kv put secret/kura/pipeline \
  webhook_secret="${WEBHOOK_SECRET}" \
  github_token="${GITHUB_TOKEN:-}"

bao kv put secret/kura/grafana \
  admin_user="${GF_SECURITY_ADMIN_USER:-admin}" \
  admin_password="${GF_SECURITY_ADMIN_PASSWORD}"

echo "✅ Secrets Kura initialisés dans Vault."
