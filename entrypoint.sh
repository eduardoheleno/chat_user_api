#!/bin/bash

set -e

# Read and set the database string
DB_PASSWORD=$(cat /run/secrets/user_database_pssw)
export GOOSE_DBSTRING="root:${DB_PASSWORD}@tcp(user_database:3306)/user_api?charset=utf8&parseTime=true"

# Debug: show the first few characters of the connection string
echo "DBSTRING: root:${DB_PASSWORD:0:3}*****@tcp(user_database:3306)/user_api?charset=utf8&parseTime=true"

exec "$@"
