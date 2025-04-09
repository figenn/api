# --- base ---
    FROM golang:1.24-alpine AS base

    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    RUN go install github.com/pressly/goose/v3/cmd/goose@v3.7.0
    
    COPY . .
    
    # --- dev ---
    FROM base AS dev
    
    CMD ["go", "run", "cmd/api/main.go"]
    
    # --- build ---
    FROM base AS build
    
    RUN go build -o main cmd/api/main.go
    
    # --- prod ---
    FROM alpine:3.20.1 AS prod
    
    WORKDIR /app
    
    RUN apk add --no-cache postgresql-client bash
    
    COPY --from=build /go/bin/goose /usr/local/bin/goose
    COPY --from=build /app/main /app/main
    COPY --from=build /app/migrations /app/migrations
    
    COPY entrypoint.sh /app/
    RUN chmod +x /app/entrypoint.sh
    
    EXPOSE 8080
    
    ENTRYPOINT ["./entrypoint.sh"]