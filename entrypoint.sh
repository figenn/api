#!/bin/bash
set -e

export BLUEPRINT_DB_HOST=postgres
export BLUEPRINT_DB_PORT=5432
export BLUEPRINT_DB_DATABASE=figenn
export BLUEPRINT_DB_USERNAME=melkey
export BLUEPRINT_DB_PASSWORD=password1234
export BLUEPRINT_DB_SCHEMA=public

DB_USER=${BLUEPRINT_DB_USERNAME}
DB_PASSWORD=${BLUEPRINT_DB_PASSWORD}
DB_HOST=${BLUEPRINT_DB_HOST}
DB_PORT=${BLUEPRINT_DB_PORT}
DB_NAME=${BLUEPRINT_DB_DATABASE}
DB_SCHEMA=${BLUEPRINT_DB_SCHEMA}

sleep 1

echo -n "🔄 Vérification de la connexion à la base de données... "
if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' >/dev/null 2>&1; then
    echo "✅ Connexion réussie!"
else
    echo "❌ Impossible de se connecter à la base de données"
    exit 1
fi

DB_STRING="user=$DB_USER password=$DB_PASSWORD host=$DB_HOST port=$DB_PORT dbname=$DB_NAME sslmode=disable search_path=$DB_SCHEMA"

echo "📋 Statut actuel des migrations:"
goose -dir ./migrations postgres "$DB_STRING" status

COMMAND=${1:-up}
echo "🚀 Exécution de la commande: goose $COMMAND"
goose -dir ./migrations postgres "$DB_STRING" "$COMMAND"
goose -dir ./migrations postgres "$DB_STRING" status

echo "✅ Migrations terminées avec succès!"

# Démarrer l'API Go
echo "🚀 Démarrage de l'API..."
/app/main
