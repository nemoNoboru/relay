# syntax=docker/dockerfile:1

# Relay node image: static site hosting + uploads, one binary, one data dir.
#
#   docker build -t relay .
#   docker run -d -p 8080:8080 -v relay-data:/data relay
#
# The image runs as an unprivileged user; /data (owned by that user in the
# image) holds sites/, the deploy token, and future node state.

# ---------- Stage 1: build the owner UI (web/dist) ----------
FROM node:22-alpine AS web
WORKDIR /src
COPY web/package.json web/package-lock.json ./web/
RUN cd web && npm ci --include=dev
COPY web/ ./web/
# copy-skill.mjs reads the canonical skill from .pi/skills at build time.
COPY .pi/skills/relay-deploy/SKILL.md ./.pi/skills/relay-deploy/SKILL.md
RUN cd web && npm run build

# ---------- Stage 2: build the relay binary ----------
FROM golang:1.26-alpine AS build
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY web/embed.go ./web/embed.go
COPY web/fonts ./web/fonts
COPY site.go host.html why.html styles.css ./
COPY --from=web /src/web/dist ./web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /relay ./cmd/relay

# ---------- Stage 3: runtime ----------
FROM alpine:3.21
# yt-dlp + ffmpeg power the publish page's "Relay a url" re-hosting; the
# node still works without them (it falls back to a share card).
RUN apk add --no-cache yt-dlp ffmpeg ca-certificates \
    && addgroup -S relay && adduser -S -G relay relay \
    && mkdir -p /data && chown relay:relay /data
COPY --from=build /relay /usr/local/bin/relay
USER relay
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["relay", "node", "--data", "/data", "--listen", ":8080"]
