FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates && update-ca-certificates

# Étape 1 : copier go.mod et go.sum (si présent)
COPY go.mod go.sum* ./

# Étape 2 : télécharger toutes les dépendances
RUN go mod download

# Étape 3 : copier le code source et s'assurer que go.sum contient bien toutes les entrées nécessaires
COPY . .
RUN go mod tidy

# Étape 4 : builder le binaire
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o terraform-service ./main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/terraform-service /app/terraform-service

EXPOSE 8082

ENV TERRAFORM_SERVICE_PORT=8082

ENTRYPOINT ["/app/terraform-service"]
