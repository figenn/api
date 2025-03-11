FROM golang:1.24-alpine AS build

WORKDIR /app

# Copier les fichiers de configuration Go
COPY go.mod go.sum ./
RUN go mod download

# Installer goose pour les migrations
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Copier le reste du code source
COPY . .

# Construire le binaire de l'application
RUN go build -o main cmd/api/main.go

FROM alpine:3.20.1 AS prod

WORKDIR /app

# Installer PostgreSQL client pour psql
RUN apk add --no-cache postgresql-client bash

# Copier goose de l'étape de build
COPY --from=build /go/bin/goose /usr/local/bin/goose

# Copier le binaire compilé et les migrations
COPY --from=build /app/main /app/main
COPY --from=build /app/migrations /app/migrations

# Exposer le port de l'application
EXPOSE ${PORT}

# Copier le script d'entrée
COPY entrypoint.sh /app/

# Assurez-vous que le fichier a les bonnes permissions d'exécution
RUN chmod +x /app/entrypoint.sh

# Utiliser bash comme interpréteur (maintenant installé)
ENTRYPOINT ["/bin/bash", "/app/entrypoint.sh"]