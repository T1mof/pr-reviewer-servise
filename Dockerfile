# Build stage
FROM golang:1.23-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -tags netgo \
    -trimpath \
    -o /app/main \
    ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /app/main /app/main
COPY --from=builder /app/migrations /app/migrations

WORKDIR /app

EXPOSE 8080

ENTRYPOINT ["/app/main"]
