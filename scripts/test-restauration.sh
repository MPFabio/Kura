#!/bin/bash
# Test de restauration au point dans le temps.
#
# Principe : on ecrit un temoin, on note l'heure, on le supprime, puis on
# restaure a l'instant d'avant la suppression sur une instance jetable. Si le
# temoin est present dans l'instance restauree, la chaine tient.
#
# La production n'est jamais touchee : la restauration se fait dans un
# repertoire temporaire, servi par un conteneur separe sur un autre port.
set -e

PSQL="sudo docker exec kura-postgres psql --username kura --dbname kura"
TRAVAIL=/tmp/restauration
sudo rm -rf "${TRAVAIL}"
sudo mkdir -p "${TRAVAIL}/pgdata" "${TRAVAIL}/wal"

echo "=== 1. ecriture du temoin ==="
${PSQL} --quiet --command "
  CREATE TABLE IF NOT EXISTS restauration_temoin (id serial primary key, note text, ecrit_a timestamptz default now());
  INSERT INTO restauration_temoin (note) VALUES ('temoin du test de restauration');
"
${PSQL} --tuples-only --command "SELECT 'temoins presents : ' || count(*) FROM restauration_temoin;"

# Le segment courant doit etre archive pour que l'insertion soit rejouable.
${PSQL} --quiet --command "SELECT pg_switch_wal();" >/dev/null
sleep 5

CIBLE=$(${PSQL} --tuples-only --no-align --command "SELECT now();")
echo "=== 2. instant cible de la restauration : ${CIBLE} ==="
sleep 3

echo "=== 3. suppression du temoin ==="
${PSQL} --quiet --command "DELETE FROM restauration_temoin;"
${PSQL} --tuples-only --command "SELECT 'temoins presents apres suppression : ' || count(*) FROM restauration_temoin;"
${PSQL} --quiet --command "SELECT pg_switch_wal();" >/dev/null
sleep 8

echo "=== 4. recuperation de la derniere sauvegarde depuis le stockage objet ==="
DERNIERE=$(sudo docker exec kura-postgres-backup sh -c "mc ls objet/kura-backup/base/ | awk '{print \$NF}' | grep 'tar.gz$' | sort | tail -n 1")
echo "sauvegarde retenue : ${DERNIERE}"
ETIQUETTE=$(echo "${DERNIERE}" | sed 's/tar\.gz$/label/')

sudo docker exec kura-postgres-backup sh -c "mc cat objet/kura-backup/base/${DERNIERE}" | sudo tee "${TRAVAIL}/base.tar.gz" >/dev/null
sudo docker exec kura-postgres-backup sh -c "mc cat objet/kura-backup/base/${ETIQUETTE}" | sudo tee "${TRAVAIL}/backup_label" >/dev/null
sudo docker exec kura-postgres-backup sh -c "cd /wal-archive && tar --create --file - ." | sudo tar --extract --file - --directory "${TRAVAIL}/wal"

echo "=== 5. reconstitution du repertoire de donnees ==="
sudo tar --extract --gzip --file "${TRAVAIL}/base.tar.gz" --directory "${TRAVAIL}/pgdata"
# backup_label indique a PostgreSQL a quel endroit du journal commencer le
# rejeu. Sans lui, il croit repartir d'un arret propre et ignore les WAL.
sudo cp "${TRAVAIL}/backup_label" "${TRAVAIL}/pgdata/backup_label"
sudo mkdir -p "${TRAVAIL}/pgdata/pg_wal"

# recovery.signal declenche le mode restauration ; sans ce fichier PostgreSQL
# demarre normalement et ignore la cible temporelle.
sudo tee "${TRAVAIL}/pgdata/postgresql.auto.conf" >/dev/null <<CONF
restore_command = 'cp /wal-archive/%f %p'
recovery_target_time = '${CIBLE}'
recovery_target_action = 'promote'
CONF
sudo touch "${TRAVAIL}/pgdata/recovery.signal"
sudo chown -R 70:70 "${TRAVAIL}/pgdata"
sudo chmod 700 "${TRAVAIL}/pgdata"

echo "=== 6. demarrage de l instance restauree ==="
sudo docker rm --force kura-restauration-test >/dev/null 2>&1 || true
sudo docker run --detach --name kura-restauration-test \
  --volume "${TRAVAIL}/pgdata:/var/lib/postgresql/data" \
  --volume "${TRAVAIL}/wal:/wal-archive:ro" \
  --env POSTGRES_PASSWORD=restauration \
  --publish 127.0.0.1:5433:5432 \
  postgres:15-alpine >/dev/null

for i in $(seq 1 30); do
  sleep 4
  if sudo docker exec kura-restauration-test pg_isready --quiet 2>/dev/null; then break; fi
done

echo "=== 7. verification du temoin dans l instance restauree ==="
sudo docker exec kura-restauration-test psql --username kura --dbname kura \
  --command "SELECT id, note, ecrit_a FROM restauration_temoin;" || {
    echo "--- journal de la restauration ---"
    sudo docker logs kura-restauration-test --tail 30
    exit 1
  }
