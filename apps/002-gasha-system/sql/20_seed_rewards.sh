#!/bin/bash
set -euo pipefail

if [ ! -f /docker-entrypoint-initdb.d/pokemonList.csv ]; then
  echo "pokemonList.csv not found"
  exit 1
fi

awk -F',' '
  NR > 1 && $1 ~ /^[0-9]+$/ {
    name = $2
    gsub(/\r$/, "", name)
    gsub(/\047/, "\047\047", name)
    printf("REPLACE INTO rewards (reward_id, name) VALUES (%d, \047%s\047);\n", $1, name)
  }
' /docker-entrypoint-initdb.d/pokemonList.csv > /tmp/seed_rewards.sql

mysql --default-character-set=utf8mb4 -uroot -p"${MYSQL_ROOT_PASSWORD}" "${MYSQL_DATABASE}" < /tmp/seed_rewards.sql
