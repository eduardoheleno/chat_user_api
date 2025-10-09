#!/bin/sh

set -e
PASS=$(cat /run/secrets/user_database_pssw)
export GOOSE_DBSTRING="root:${PASS}@tcp(user_database:3306)/user_api?charset=utf8&parseTime=true"
exec "$@"
