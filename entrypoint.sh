#!/bin/bash
set -e

if [ -f ".env" ]; then
  set -a
  source .env
  set +a
fi

DB_USER=${BLUEPRINT_DB_USERNAME}
DB_PASSWORD=${BLUEPRINT_DB_PASSWORD}
DB_HOST=${BLUEPRINT_DB_HOST}
DB_PORT=${BLUEPRINT_DB_PORT}
DB_NAME=${BLUEPRINT_DB_DATABASE}
DB_SCHEMA=${BLUEPRINT_DB_SCHEMA}

echo "🔄 Attente de PostgreSQL ($DB_HOST:$DB_PORT)..."
until PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' >/dev/null 2>&1; do
  echo "⏳ PostgreSQL pas encore prêt..."
  sleep 2
done
echo "✅ PostgreSQL est prêt"

DB_STRING="user=$DB_USER password=$DB_PASSWORD host=$DB_HOST port=$DB_PORT dbname=$DB_NAME sslmode=disable search_path=$DB_SCHEMA"

echo "📋 Goose status:"
goose -dir ./migrations postgres "$DB_STRING" status

COMMAND=${1:-up}
echo "🚀 goose $COMMAND"
goose -dir ./migrations postgres "$DB_STRING" "$COMMAND"

exec /app/main