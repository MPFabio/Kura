#!/bin/sh
# Sauvegarde PostgreSQL vers le stockage objet.
#
# Deux flux, de frequences differentes, parce qu'ils ne repondent pas a la
# meme question :
#   - les segments WAL partent en continu, ils bornent la perte de donnees ;
#   - la sauvegarde de base part chaque nuit, elle donne le point de depart a
#     partir duquel les WAL sont rejoues.
#
# Une restauration au point dans le temps a besoin des deux. Une sauvegarde de
# base sans WAL ramene a la nuit precedente ; des WAL sans sauvegarde de base
# ne se rejouent sur rien.
#
# Note sur les options : l'image est basee sur Alpine, dont busybox n'accepte
# pas les options longues de wc, head, read et rm. Elles sont donc en forme courte
# ici, contrairement au reste du depot.
set -e

INTERVALLE_WAL="${INTERVALLE_WAL:-300}"
INTERVALLE_BASE="${INTERVALLE_BASE:-86400}"
RETENTION_BASES="${RETENTION_BASES:-3}"
# Duree de conservation des segments WAL, en heures. Elle doit couvrir la
# plus ancienne sauvegarde de base conservee : un WAL supprime avant elle
# rend cette sauvegarde irrejouable, donc inutile.
RETENTION_WAL_HEURES="${RETENTION_WAL_HEURES:-96}"
DEPOT="${DEPOT_OBJET:-kura-backup}"

export PGHOST="${PGHOST:-postgres}"
export PGUSER="${POSTGRES_USER:-kura}"
export PGPASSWORD="${POSTGRES_PASSWORD}"
export PGDATABASE="${POSTGRES_DB:-kura}"

journal() { echo "$(date --utc '+%Y-%m-%dT%H:%M:%SZ') $*"; }

journal "attente de la base"
until pg_isready --quiet; do sleep 3; done

mc alias set objet "http://minio:9000" "${MINIO_ROOT_USER}" "${MINIO_ROOT_PASSWORD}" >/dev/null
mc mb --ignore-existing "objet/${DEPOT}" >/dev/null
journal "depot objet pret : ${DEPOT}"

rotation() {
  total=$(mc ls "objet/${DEPOT}/base/" 2>/dev/null | awk '{print $NF}' | grep -c 'tar.gz$' || true)
  surplus=$((total - RETENTION_BASES))
  [ "${surplus}" -gt 0 ] || return 0

  mc ls "objet/${DEPOT}/base/" | awk '{print $NF}' | grep 'tar.gz$' | sort | head -n "${surplus}" \
    | while read -r fichier; do
        journal "rotation : suppression de ${fichier}"
        mc rm "objet/${DEPOT}/base/${fichier}" >/dev/null
        etiquette=$(echo "${fichier}" | sed 's/tar\.gz$/label/')
        mc rm "objet/${DEPOT}/base/${etiquette}" >/dev/null 2>&1 || true
      done
}

purge_wal() {
  # Purge locale : un segment n'est supprime qu'apres avoir ete depose sur le
  # stockage objet, la synchronisation venant de reussir juste au-dessus.
  # Sans cette purge, /wal-archive croit sans limite et finit par saturer le
  # disque de la machine, ce qui arrete PostgreSQL avant d'arreter la
  # sauvegarde.
  supprimes=$(find /wal-archive -type f -mmin +$((RETENTION_WAL_HEURES * 60)) -delete -print 2>/dev/null | wc -l)
  [ "${supprimes}" -gt 0 ] && journal "purge locale : ${supprimes} segments retires"

  # Purge cote objet : meme regle, appliquee a la copie distante.
  anciens=$(mc find "objet/${DEPOT}/wal" --older-than "${RETENTION_WAL_HEURES}h" 2>/dev/null | wc -l)
  if [ "${anciens}" -gt 0 ]; then
    mc find "objet/${DEPOT}/wal" --older-than "${RETENTION_WAL_HEURES}h" --exec "mc rm {}" >/dev/null 2>&1 || true
    journal "purge objet : ${anciens} segments de plus de ${RETENTION_WAL_HEURES} h retires"
  fi
  return 0
}

sauvegarde_de_base() {
  horodatage=$(date --utc '+%Y%m%dT%H%M%SZ')
  journal "sauvegarde de base ${horodatage} : debut"

  # pg_backup_start et pg_backup_stop encadrent la copie et doivent etre
  # appeles dans la meme session : c'est elle qui tient la sauvegarde ouverte.
  # \! execute la copie sans quitter cette session.
  #
  # pg_wal est exclu : les segments courants sont deja archives par
  # archive_command, les inclure ferait doublon et fausserait la restauration.
  # --tuples-only et --no-align : le fichier d'etiquette doit contenir le
  # contenu brut renvoye par pg_backup_stop. Avec la mise en forme par
  # defaut, psql y ajoute les en-tetes de colonnes et le compteur de lignes,
  # et PostgreSQL refuse la restauration sur un backup_label invalide.
  psql --quiet --no-psqlrc --tuples-only --no-align <<SQL
SELECT pg_backup_start('kura-${horodatage}', true);
\\! tar --create --gzip --file /travail/base-${horodatage}.tar.gz --directory /pgdata --exclude=pg_wal/* .
SELECT labelfile FROM pg_backup_stop(true) \\g /travail/base-${horodatage}.label
SQL

  mc cp --quiet "/travail/base-${horodatage}.tar.gz" "objet/${DEPOT}/base/" >/dev/null
  mc cp --quiet "/travail/base-${horodatage}.label" "objet/${DEPOT}/base/" >/dev/null
  taille=$(du -h "/travail/base-${horodatage}.tar.gz" | awk '{print $1}')
  rm -f "/travail/base-${horodatage}.tar.gz" "/travail/base-${horodatage}.label"

  rotation
  journal "sauvegarde de base ${horodatage} : terminee (${taille})"
}

# Premiere sauvegarde immediate : sans elle, les WAL archives jusqu'a la
# premiere nuit ne seraient rejouables sur rien.
sauvegarde_de_base
derniere_base=$(date +%s)

while true; do
  # mirror plutot que cp : seuls les segments nouveaux traversent le reseau.
  if mc mirror --quiet --overwrite /wal-archive "objet/${DEPOT}/wal" >/dev/null 2>&1; then
    segments=$(ls -1 /wal-archive 2>/dev/null | wc -l)
    journal "WAL synchronises (${segments} segments locaux)"
  else
    journal "ECHEC de la synchronisation des WAL"
  fi

  purge_wal

  maintenant=$(date +%s)
  if [ $((maintenant - derniere_base)) -ge "${INTERVALLE_BASE}" ]; then
    sauvegarde_de_base
    derniere_base=${maintenant}
  fi

  sleep "${INTERVALLE_WAL}"
done
