# Relay backend

The Python (FastAPI) backend of Relay. It replaces the old Go node, and is
built as a SOLID + Domain-Driven Design skeleton: one package
per bounded context, with `api`, `application`, `domain` and `infrastructure`
layers. Today it ships the skeleton and `GET /health`.

The architecture, testing plan and tooling are in
[`docs/backend-development.md`](../docs/backend-development.md).

## Prerequisites

[uv](https://docs.astral.sh/uv/) only. It installs Python (see `.python-version`),
the virtualenv and the locked dependencies.

## Quickstart

Run everything from `backend/`:

```sh
uv sync                                  # install Python + dependencies
uv run alembic upgrade head              # create/upgrade the SQLite database
uv run uvicorn relay.main:app --reload   # serve on http://localhost:8000
curl localhost:8000/health               # {"status":"ok","version":"dev"}
uv run pytest                            # unit, integration, API tests
```

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `DATABASE_URL` | `sqlite+aiosqlite:///./relay.db` | SQLAlchemy URL. `postgresql+asyncpg://user:pass@host/db` switches to Postgres, no code change |
| `VERSION` | `dev` | Reported by `/health` |
| `LOG_LEVEL` | `INFO` | JSON logs on stdout |

## Checks

```sh
uv run ruff check && uv run ruff format --check
uv run ty check
uv run lint-imports
uv run pytest
```
