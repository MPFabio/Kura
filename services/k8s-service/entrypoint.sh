#!/bin/sh
# Si GCP_SA_KEY_JSON est défini (ex. en SaaS, injecté par un secret manager),
# on écrit la clé dans un fichier pour que le plugin GKE puisse l'utiliser.
if [ -n "$GCP_SA_KEY_JSON" ]; then
  echo "$GCP_SA_KEY_JSON" > /tmp/gcp-sa.json
  export GOOGLE_APPLICATION_CREDENTIALS=/tmp/gcp-sa.json
fi
exec /app/k8s-service "$@"
