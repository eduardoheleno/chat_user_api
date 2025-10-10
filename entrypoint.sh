#!/bin/bash

set -e

DB_PASSWORD=$(cat /run/secrets/user_database_pssw)
export GOOSE_DBSTRING="root:${DB_PASSWORD}@tcp(user_database:3306)/user_api?charset=utf8&parseTime=true"

exec "$@"
