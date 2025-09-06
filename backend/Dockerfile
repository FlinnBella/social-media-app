# Multi-stage build for the Go API with ffmpeg in the runtime image

FROM golang:1.21-bookworm AS builder
WORKDIR /app/backend

# Cache deps
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source
COPY backend/ .

# Build static binary
ENV CGO_ENABLED=0
RUN go build -o /out/server main.go


FROM debian:bookworm-slim AS runtime
WORKDIR /app/backend

# System deps: ffmpeg for video composition, certificates, tzdata
RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    ffmpeg \
    ca-certificates \
    tzdata \
  && rm -rf /var/lib/apt/lists/*

# Copy application binary and assets
COPY --from=builder /out/server /usr/local/bin/server
COPY backend/music ./music

ENV APP_ENV=production
ENV PORT=8080

EXPOSE 8080

CMD ["server"]


