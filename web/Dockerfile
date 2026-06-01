# Multi-stage build: node compiles the Vite bundle, nginx serves the
# static output. Final image is ~50MB (alpine + nginx + static assets).
#
# VITE_API_URL is empty so the built bundle uses relative `/api/...`
# paths — the SPA + API share an origin in production (kerbecs handles
# routing), so relative URLs are the cleanest setup.

FROM --platform=$BUILDPLATFORM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . ./

# Empty VITE_API_URL → axios baseURL becomes "/api" → same-origin requests.
ENV VITE_API_URL=""
RUN npm run build

##
## Deploy
##
FROM nginx:1.27-alpine

# SPA-aware nginx config: try_files falls back to index.html so client-side
# router URLs (e.g. /groups/xxx) resolve on full-page refresh.
COPY nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s \
  CMD wget -q --spider http://localhost/ || exit 1
