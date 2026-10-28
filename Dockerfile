# syntax=docker/dockerfile:1

# Relay: the FastAPI backend serving the Expo web build, one image, one port.
#
#   docker build -t relay .
#   docker run -d -p 8000:8000 -v relay-data:/data relay
#
# http://localhost:8000/ is the web app, /api/* the API (/api/health for probes).
# Runs as an unprivileged user (uid/gid 10001). /data holds the SQLite database;
# set DATABASE_URL=postgresql+asyncpg://... to use Postgres instead.

# ---------- Stage 1: export the Expo web build ----------
FROM node:22-slim AS web
WORKDIR /src/mobile
COPY mobile/package.json mobile/package-lock.json ./
RUN npm ci
COPY mobile/ ./
RUN npx expo export --platform web --output-dir /web-dist

# ---------- Stage 2: resolve and install locked backend dependencies ----------
FROM ghcr.io/astral-sh/uv:python3.13-bookworm-slim AS build
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy UV_PYTHON_DOWNLOADS=never
WORKDIR /app

# Dependencies first so this layer is cached until the lockfile changes.
COPY backend/pyproject.toml backend/uv.lock backend/.python-version backend/README.md ./
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev --no-install-project

COPY backend/src/ ./src/
COPY backend/alembic.ini ./
COPY backend/migrations/ ./migrations/
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev

# ---------- Stage 3: runtime ----------
FROM python:3.13-slim-bookworm
ARG VERSION=dev
ENV VERSION=${VERSION} \
    PATH="/app/.venv/bin:$PATH" \
    PYTHONUNBUFFERED=1 \
    DATABASE_URL=sqlite+aiosqlite:////data/relay.db \
    WEB_DIST_DIR=/app/web

RUN groupadd --system --gid 10001 relay \
    && useradd --system --uid 10001 --gid relay --home-dir /app relay \
    && mkdir /data && chown relay:relay /data
WORKDIR /app
COPY --from=build --chown=relay:relay /app /app
COPY --from=web --chown=relay:relay /web-dist /app/web

USER relay
VOLUME /data
EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["python", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/api/health', timeout=2)"]

CMD ["uvicorn", "relay.main:app", "--host", "0.0.0.0", "--port", "8000"]
