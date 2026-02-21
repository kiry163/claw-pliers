FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o claw-pliers ./cmd/claw-pliers/

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/claw-pliers .
COPY data/config/config.yaml ./config.yaml
COPY data/config/*.yaml ./config/

RUN mkdir -p /app/data

EXPOSE 8080

ENV CLAWPLIERS_SERVER_PORT=8080
ENV CLAWPLIERS_DATABASE_PATH=/app/data/claw-pliers.db

ENTRYPOINT ["./claw-pliers"]
