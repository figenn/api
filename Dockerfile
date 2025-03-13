FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY . .

RUN go build -o main cmd/api/main.go

FROM alpine:3.20.1 AS prod

WORKDIR /app

RUN apk add --no-cache postgresql-client bash

COPY --from=build /go/bin/goose /usr/local/bin/goose

COPY --from=build /app/main /app/main
COPY --from=build /app/migrations /app/migrations

EXPOSE ${PORT}

COPY entrypoint.sh /app/

RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/bin/bash", "/app/entrypoint.sh"]