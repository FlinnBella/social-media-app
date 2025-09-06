# Build the Vite frontend and serve via Caddy

FROM node:20-bullseye AS frontend
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build


FROM caddy:2.8-alpine
WORKDIR /srv

# Copy built frontend to Caddy web root
COPY --from=frontend /app/dist /srv

# Use container-specific Caddyfile
COPY caddy/Caddyfile /etc/caddy/Caddyfile

EXPOSE 80 443


