# --- Build stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

# --- Run stage ---
FROM alpine:3.20

WORKDIR /app
COPY --from=builder /server /app/server
COPY configs/application.properties /app/configs/application.properties

EXPOSE 8080
ENTRYPOINT ["/app/server"]
