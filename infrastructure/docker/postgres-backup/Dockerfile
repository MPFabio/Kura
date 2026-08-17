# Image de sauvegarde : les outils client PostgreSQL et le client MinIO.
# Elle ne contient pas de serveur, seulement de quoi lire la base et deposer
# les archives sur le stockage objet.
FROM postgres:15-alpine

ARG MC_VERSION=RELEASE.2024-01-13T08-44-48Z

RUN apk add --no-cache curl tar gzip \
    && curl --silent --show-error --fail --location \
       "https://dl.min.io/client/mc/release/linux-amd64/archive/mc.${MC_VERSION}" \
       --output /usr/local/bin/mc \
    && chmod +x /usr/local/bin/mc

COPY sauvegarde.sh /usr/local/bin/sauvegarde.sh
RUN chmod +x /usr/local/bin/sauvegarde.sh

ENTRYPOINT ["/usr/local/bin/sauvegarde.sh"]
