# --- ETAPA 1: Compilación (Usando la versión más reciente de Go) ---
FROM golang:alpine AS builder

WORKDIR /app

# Copiar archivos de módulos
COPY go.mod go.sum ./
RUN go mod download

# Copiar todo el código fuente
COPY . .

# Compilar binario estático ligero
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o cloudsentinel .

# --- ETAPA 2: Imagen Final Minimalista de Producción ---
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copiar únicamente el binario compilado
COPY --from=builder /app/cloudsentinel .

EXPOSE 8080

ENTRYPOINT ["./cloudsentinel"]
