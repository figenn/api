#!/bin/bash
set -e

set -a
source .env
set +a

DB_USER=${BLUEPRINT_DB_USERNAME}
DB_PASSWORD=${BLUEPRINT_DB_PASSWORD}
DB_HOST="postgres"
DB_PORT=${BLUEPRINT_DB_PORT}
DB_NAME=${BLUEPRINT_DB_DATABASE}
DB_SCHEMA=${BLUEPRINT_DB_SCHEMA}

export SMTP_HOST= "mailhog"

echo "🔄 Attente de la base de données PostgreSQL ($DB_HOST:$DB_PORT)..."
until PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' >/dev/null 2>&1; do
  echo "⏳ En attente que PostgreSQL soit prêt..."
  sleep 2
done
echo "✅ Connexion à PostgreSQL réussie !"

DB_STRING="user=$DB_USER password=$DB_PASSWORD host=$DB_HOST port=$DB_PORT dbname=$DB_NAME sslmode=disable search_path=$DB_SCHEMA"

echo "📋 Statut actuel des migrations:"
goose -dir ./migrations postgres "$DB_STRING" status

COMMAND=${1:-up}
echo "🚀 Exécution de la commande: goose $COMMAND"
goose -dir ./migrations postgres "$DB_STRING" "$COMMAND"
goose -dir ./migrations postgres "$DB_STRING" status

echo "✅ Migrations terminées avec succès!"

echo "🚀 Démarrage de l'API..."
exec /app/main