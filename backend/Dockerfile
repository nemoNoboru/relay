# syntax=docker/dockerfile:1

# Relay backend image (FastAPI). Build context is backend/:
#
#   docker build -t relay-backend --build-arg VERSION=dev backend
#   docker run -d -p 8000:8000 -v relay-backend-data:/data relay-backend
#
# Runs as an unprivileged user. /data holds the default SQLite database;
# set DATABASE_URL=postgresql+asyncpg://... to use Postgres instead.

# ---------- Stage 1: resolve and install locked dependencies ----------
FROM ghcr.io/astral-sh/uv:python3.13-bookworm-slim AS build
ENV UV_COMPILE_BYTECODE=1 UV_LINK_MODE=copy UV_PYTHON_DOWNLOADS=never
WORKDIR /app

# Dependencies first so this layer is cached until the lockfile changes.
COPY pyproject.toml uv.lock .python-version README.md ./
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev --no-install-project

COPY src/ ./src/
COPY alembic.ini ./
COPY migrations/ ./migrations/
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev

# ---------- Stage 2: runtime ----------
FROM python:3.13-slim-bookworm
ARG VERSION=dev
ENV VERSION=${VERSION} \
    PATH="/app/.venv/bin:$PATH" \
    PYTHONUNBUFFERED=1 \
    DATABASE_URL=sqlite+aiosqlite:////data/relay.db

RUN useradd --system --uid 10001 --home-dir /app relay \
    && mkdir /data && chown relay /data
WORKDIR /app
COPY --from=build --chown=relay /app /app

USER relay
VOLUME /data
EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["python", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/health', timeout=2)"]

CMD ["uvicorn", "relay.main:app", "--host", "0.0.0.0", "--port", "8000"]
