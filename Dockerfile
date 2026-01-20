FROM golang:1.25-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o ddownloader ./cmd/ddownloader/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/ddownloader .

COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

EXPOSE 3000

# Ejecutamos la aplicación
CMD ["./ddownloader"]