# Build the Vite frontend and serve via Caddy

FROM node:20-bullseye AS frontend
WORKDIR /app

# 1) Install deps using only lockfiles for caching
COPY package.json package-lock.json ./
RUN npm ci

# 2) Copy only files needed for Vite build to maximize cache hits
COPY index.html ./
COPY vite.config.* ./
COPY tsconfig*.json ./
COPY postcss.config.js ./
COPY tailwind.config.js ./
COPY src ./src

# 3) Build the frontend
RUN npm run build


FROM caddy:2.8-alpine
WORKDIR /srv

# Copy built frontend to Caddy web root
COPY --from=frontend /app/dist /srv

# Use container-specific Caddyfile
COPY caddy/Caddyfile /etc/caddy/Caddyfile

EXPOSE 80 443


