# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ansible-service ./main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /root/

COPY --from=builder /app/ansible-service .

EXPOSE 8083

CMD ["./ansible-service"]
