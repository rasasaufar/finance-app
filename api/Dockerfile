FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN GOTOOLCHAIN=auto go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOTOOLCHAIN=auto go build -o api ./cmd/api/

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/api .
RUN mkdir -p /app/uploads
EXPOSE 8080
CMD ["./api"]
