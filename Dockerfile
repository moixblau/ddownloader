# ETAPA 1: Construcción (Build)
FROM golang:1.25-alpine AS builder

# Instalamos certificados por si tu app hace peticiones HTTPS
RUN apk add --no-cache ca-certificates git

# Definimos el directorio de trabajo
WORKDIR /app

# Copiamos archivos de dependencias primero para aprovechar la caché de Docker
COPY go.mod go.sum ./
RUN go mod download

# Copiamos el resto del código
COPY . .

# Compilamos el binario (CGO_ENABLED=0 genera un binario estático que corre en cualquier Linux)
RUN CGO_ENABLED=0 GOOS=linux go build -o ddownloader main.go

# ETAPA 2: Imagen Final (Producción)
FROM alpine:latest

# Instalamos ca-certificates de nuevo para llamadas API seguras
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiamos el binario desde la etapa de construcción
COPY --from=builder /app/ddownloader .

# Copiamos carpetas estáticas y templates (Muy importante para tu proyecto)
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Exponemos el puerto (ajusta el 8080 si usas otro)
EXPOSE 8080

# Ejecutamos la aplicación
CMD ["./ddownloader"]