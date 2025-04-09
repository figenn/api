#!/bin/bash
# Charger les variables depuis le fichier .env s'il existe
if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

DB_USER=${BLUEPRINT_DB_USERNAME:?Missing DB username}
DB_PASSWORD=${BLUEPRINT_DB_PASSWORD:?Missing DB password}
DB_HOST=${BLUEPRINT_DB_HOST:?Missing DB host}
DB_PORT=${BLUEPRINT_DB_PORT:-5432}
DB_NAME=${BLUEPRINT_DB_DATABASE:?Missing DB name}
DB_SCHEMA=${BLUEPRINT_DB_SCHEMA:-public}

echo "🔧 Configuration de la base de données:"
echo "  - Host: $DB_HOST:$DB_PORT"
echo "  - Database: $DB_NAME"
echo "  - User: $DB_USER"
echo "  - Schema: $DB_SCHEMA"

echo -n "🔄 Vérification de la connexion à la base de données... "
if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' >/dev/null 2>&1; then
    echo "✅ Connexion réussie!"
else
    echo "❌ Impossible de se connecter à la base de données"
    exit 1
fi

DB_STRING="user=$DB_USER password=$DB_PASSWORD host=$DB_HOST port=$DB_PORT dbname=$DB_NAME sslmode=disable search_path=$DB_SCHEMA"

echo "📊 Version de Goose:"
goose -version

echo "📋 Statut actuel des migrations:"
goose -dir ./migrations postgres "$DB_STRING" status

COMMAND=${1:-up}
echo "🚀 Exécution de la commande: goose $COMMAND"
goose -dir ./migrations postgres "$DB_STRING" "$COMMAND"

echo "📋 Nouveau statut des migrations:"
goose -dir ./migrations postgres "$DB_STRING" status

echo "✅ Opération terminée avec succès!"