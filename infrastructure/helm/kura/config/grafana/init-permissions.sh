#!/bin/sh
# Accorde au rôle Viewer le droit de lecture sur les tableaux de bord
# provisionnés, afin que l'iframe embarquée dans Kura (accès anonyme) puisse
# les afficher.
#
# Pourquoi ce script existe : dans Grafana (vérifié en 10.2 et en 11.3),
# l'utilisateur anonyme rattaché au rôle Viewer ne reçoit aucune permission
# sur les tableaux de bord, et ceux provisionnés par fichier naissent sans
# ACL. L'iframe reçoit alors « 403 dashboards:read » alors que la
# configuration d'authentification anonyme est correcte. Le provisioning par
# fichier ne permet pas de déclarer ces ACL : seule l'API le permet.
#
# Le script est idempotent : le rejouer ne change rien, et il traite les
# tableaux de bord découverts dynamiquement, sans liste figée.
set -eu

GRAFANA_URL="${GRAFANA_URL:-http://grafana:3000}"
ADMIN_USER="${GF_SECURITY_ADMIN_USER:-admin}"
ADMIN_PASSWORD="${GF_SECURITY_ADMIN_PASSWORD:?mot de passe admin Grafana requis}"

echo "[grafana-init] attente de Grafana sur ${GRAFANA_URL}"
i=0
until curl -sf -o /dev/null "${GRAFANA_URL}/api/health"; do
  i=$((i + 1))
  if [ "$i" -gt 60 ]; then
    echo "[grafana-init] Grafana injoignable après 5 minutes, abandon" >&2
    exit 1
  fi
  sleep 5
done

# Laisse le provisioning par fichier créer les tableaux de bord avant de
# poser les ACL : sinon la liste est vide au premier démarrage.
sleep 10

uids=$(curl -sf -u "${ADMIN_USER}:${ADMIN_PASSWORD}" \
  "${GRAFANA_URL}/api/search?type=dash-db" \
  | tr ',' '\n' | sed -n 's/.*"uid":"\([^"]*\)".*/\1/p')

if [ -z "$uids" ]; then
  echo "[grafana-init] aucun tableau de bord trouvé, rien à faire"
  exit 0
fi

for uid in $uids; do
  code=$(curl -s -o /dev/null -w '%{http_code}' -X POST \
    -u "${ADMIN_USER}:${ADMIN_PASSWORD}" \
    -H 'Content-Type: application/json' \
    -d '{"items":[{"role":"Viewer","permission":1}]}' \
    "${GRAFANA_URL}/api/dashboards/uid/${uid}/permissions")
  echo "[grafana-init] ${uid} -> HTTP ${code}"
done

echo "[grafana-init] terminé"
