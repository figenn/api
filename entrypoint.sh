#!/bin/bash
set -e

# Définir les variables d'environnement directement dans le script
export BLUEPRINT_DB_HOST=postgres
export BLUEPRINT_DB_PORT=5432
export BLUEPRINT_DB_DATABASE=figenn
export BLUEPRINT_DB_USERNAME=melkey
export BLUEPRINT_DB_PASSWORD=password1234
export BLUEPRINT_DB_SCHEMA=public

# Variables de configuration de la base de données
DB_USER=${BLUEPRINT_DB_USERNAME}
DB_PASSWORD=${BLUEPRINT_DB_PASSWORD}
DB_HOST=${BLUEPRINT_DB_HOST}
DB_PORT=${BLUEPRINT_DB_PORT}
DB_NAME=${BLUEPRINT_DB_DATABASE}
DB_SCHEMA=${BLUEPRINT_DB_SCHEMA}

# Afficher la configuration
echo "🔧 Configuration de la base de données:"
echo "  - Host: $DB_HOST:$DB_PORT"
echo "  - Database: $DB_NAME"
echo "  - User: $DB_USER"
echo "  - Schema: $DB_SCHEMA"

sleep 2

# Vérifier si la base de données est disponible
echo -n "🔄 Vérification de la connexion à la base de données... "
if PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' >/dev/null 2>&1; then
    echo "✅ Connexion réussie!"
else
    echo "❌ Impossible de se connecter à la base de données"
    exit 1
fi

# Construire la chaîne de connexion
DB_STRING="user=$DB_USER password=$DB_PASSWORD host=$DB_HOST port=$DB_PORT dbname=$DB_NAME sslmode=disable search_path=$DB_SCHEMA"

# Afficher la version de Goose
echo "📊 Version de Goose:"
goose -version

# Vérifier le statut des migrations
echo "📋 Statut actuel des migrations:"
goose -dir ./migrations postgres "$DB_STRING" status

# Exécuter les migrations selon la commande fournie
COMMAND=${1:-up}
echo "🚀 Exécution de la commande: goose $COMMAND"
goose -dir ./migrations postgres "$DB_STRING" "$COMMAND"

# Afficher le nouveau statut après exécution
echo "📋 Nouveau statut des migrations:"
goose -dir ./migrations postgres "$DB_STRING" status

echo "✅ Migrations terminées avec succès!"

# Démarrer l'API Go
echo "🚀 Démarrage de l'API..."
/app/main
