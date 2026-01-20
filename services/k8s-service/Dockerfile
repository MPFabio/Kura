FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Étape 1 : préparer le cache des dépendances à partir du go.mod
COPY go.mod ./
RUN go mod download && go mod verify

# Étape 2 : copier le code source et s'assurer que go.sum contient bien toutes les entrées nécessaires
COPY . .
RUN go mod tidy

# Étape 3 : builder le binaire
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o k8s-service ./main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/k8s-service /app/k8s-service

EXPOSE 8081

ENV K8S_SERVICE_PORT=8081

ENTRYPOINT ["/app/k8s-service"]

