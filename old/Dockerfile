FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /src

# Cache de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar código común
COPY api/ ./api/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Workspace (si existe)
COPY go.work* ./
RUN if [ -f go.work ]; then go work sync; fi

# Variables build
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
ENV GOFLAGS="-trimpath"
ENV LDFLAGS="-s -w"

# Copiar cmds y construir en paralelo
COPY cmd ./cmd
RUN mkdir -p /out && \
    go build -o /out/api         -tags netgo -ldflags "${LDFLAGS}" ./cmd/api         & \
    go build -o /out/scheduler   -tags netgo -ldflags "${LDFLAGS}" ./cmd/scheduler   & \
    go build -o /out/migrations  -tags netgo -ldflags "${LDFLAGS}" ./cmd/migrations  & \
    go build -o /out/repurposer  -tags netgo -ldflags "${LDFLAGS}" ./cmd/repurposer  & \
    go build -o /out/reconcile   -tags netgo -ldflags "${LDFLAGS}" ./cmd/reconcile   & \
    wait

# Imagen final
FROM alpine:latest

# CRITICAL: Install fonts for proper character rendering
# font-noto provides Noto Sans (for text/control chars) and basic emoji support
# font-noto-emoji provides additional emoji rendering capabilities
# Noto Sans handles missing glyphs from Montserrat (like newlines \n = U+000A)
# CAMBIO PRINCIPAL: Usar ffmpeg6 en lugar de ffmpeg
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    ffmpeg6 \
    curl \
    fontconfig \
    font-noto \
    font-noto-emoji && \
    update-ca-certificates && \
    mkdir -p /usr/share/fonts/truetype/custom

# Copy custom Montserrat fonts
COPY fonts/*.ttf /usr/share/fonts/truetype/custom/

# Copy fontconfig configuration for text and emoji fallback
# This configuration ensures Montserrat falls back to Noto Sans for missing glyphs
# (like newlines, control characters) before falling back to emoji fonts
COPY fonts/local.conf /etc/fonts/local.conf

# Update font cache and verify installation
RUN fc-cache -f -v && \
    echo "=== Installed Montserrat variants ===" && \
    fc-list | grep -i montserrat && \
    echo "" && \
    echo "=== Installed Emoji fonts ===" && \
    fc-list | grep -i "noto.*emoji" && \
    echo "" && \
    echo "=== Font configuration test ===" && \
    fc-match "Montserrat SemiBold" && \
    fc-match "emoji"

WORKDIR /app

COPY --from=builder /out/api ./api
COPY --from=builder /out/scheduler ./scheduler
COPY --from=builder /out/migrations ./migrations
COPY --from=builder /out/repurposer ./repurposer
COPY --from=builder /out/reconcile ./reconcile

EXPOSE 8080