# ---- Build stage ----
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o mirrorverse main.go

# ---- Final stage (scratch) ----
FROM scratch
COPY --from=builder /app/mirrorverse /mirrorverse
ENTRYPOINT ["/mirrorverse"]
