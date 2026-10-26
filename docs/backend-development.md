# Backend development

How the Python backend in [`backend/`](../backend/) is built and how to extend
it. It is the start of a FastAPI rewrite of the Go node; today it ships the
skeleton and `GET /health`. Quickstart: [`backend/README.md`](../backend/README.md).

Everything runs through [uv](https://docs.astral.sh/uv/): `uv sync`,
`uv run pytest`, `uv run alembic upgrade head`,
`uv run uvicorn relay.main:app`. There is no `requirements.txt` and no pip.
Pydantic v2 is the modelling basis: settings, domain entities and value
objects, application DTOs and API schemas.

## Architecture

DDD / clean architecture, one package per bounded context. Dependencies point
inward:

```
api  →  application  →  domain  ←  infrastructure
```

| Layer | Holds | May import |
|---|---|---|
| `domain` | Entities, value objects, domain services, repository interfaces (`Protocol`) | Pydantic, stdlib. Never FastAPI or SQLAlchemy |
| `application` | Use cases, DTOs (Pydantic), unit-of-work and clock `Protocol`s | domain. Never FastAPI or SQLAlchemy |
| `infrastructure` | SQLAlchemy ORM models, repository and unit-of-work implementations, external clients | domain, application. ORM models never leave this layer |
| `api` | FastAPI routers, request/response schemas, `Depends` providers. Thin | application (and the container) |

Cross-cutting pieces live in `relay.shared`: `config` (Settings), `logging`
(JSON), and shared base classes in each layer package. The layer rules are
enforced by import-linter contracts in `backend/pyproject.toml`
(`uv run lint-imports`), so a wrong-direction import fails CI.

### Adding a bounded context

Create `backend/src/relay/contexts/<name>/` with the same four packages:

```
contexts/posts/
  domain/          # Post entity, PostRepository Protocol
  application/     # CreatePost use case, PostDTO, PostsUnitOfWork Protocol
  infrastructure/  # PostModel (ORM), SqlAlchemyPostRepository, unit of work
  api/             # router, request/response schemas, providers
```

Then: add the layers to the import-linter contracts, import the ORM models in
`migrations/env.py` (so autogenerate sees them), create a migration
(`uv run alembic revision --autogenerate -m "posts"`), add a `get_<use_case>()`
method to the `Container`, and include the router in `create_app`.

## Dependency injection

Constructor injection throughout.

- Every service, use case, repository and unit of work receives its
  collaborators in `__init__`. No module-level singletons, no service locator,
  no `Settings()` inside classes.
- Classes depend on `Protocol`s from domain/application, never on concrete
  infrastructure classes.
- `relay/main.py` is the single composition root. `create_app(settings)` takes
  config as an argument; the lifespan builds the engine and session factory
  once, stores a `Container` (plain class, `relay/container.py`) on
  `app.state`, and disposes the engine on shutdown. The module-level `app`
  that uvicorn serves is `create_app(Settings())`.
- FastAPI `Depends` appears only in `api/` packages, as thin providers that ask
  the container for a built use case.

```python
class CreatePost:
    def __init__(self, uow_factory: UnitOfWorkFactory, clock: Clock) -> None: ...
    async def __call__(self, cmd: CreatePostCommand) -> PostDTO: ...
```

## Testing

`uv run pytest` (pytest-asyncio in auto mode). Shared fixtures are in
`tests/conftest.py`, fakes in `tests/fakes/`.

| Folder | What | How |
|---|---|---|
| `tests/unit/` | Domain and application | No DB, no FastAPI. In-memory fakes (`FakeUnitOfWork`, `InMemory<X>Repository`, `FixedClock`) passed to constructors |
| `tests/integration/` | Repositories and unit of work against a real DB | SQLite by default; the same tests run on Postgres when `TEST_DATABASE_URL` is set. The schema comes from `alembic upgrade head`, never `create_all` |
| `tests/api/` | HTTP behaviour | `httpx.AsyncClient` + `ASGITransport` against `create_app(test_settings)`. Override providers with `app.dependency_overrides` |
| import-linter | Architecture | Contracts in `pyproject.toml` (`uv run lint-imports`) |

To run the Postgres paths locally:

```sh
docker run -d --name relay-pg -e POSTGRES_USER=relay -e POSTGRES_PASSWORD=relay \
  -e POSTGRES_DB=relay_test -p 5432:5432 postgres:17-alpine
TEST_DATABASE_URL=postgresql+asyncpg://relay:relay@localhost:5432/relay_test uv run pytest
```

## Lint, types, pre-commit

```sh
uv run ruff check            # E, W, F, I, B, UP, SIM, ASYNC, S, N, RUF
uv run ruff format --check   # the only formatter (line length 100)
uv run ty check              # src/ and tests/; no mypy
uv run lint-imports
```

Config is in `backend/pyproject.toml`. `S101` (assert) is allowed under
`tests/`.

The root `.pre-commit-config.yaml` runs these for files under `backend/`
(ruff check `--fix`, ruff format, `uv lock`, ty, lint-imports) next to the Go
hooks:

```sh
uv tool install pre-commit
pre-commit install
pre-commit run --all-files
```

Change dependencies with `uv add <pkg>` (or `uv add --dev <pkg>`); commit
`uv.lock`. CI installs with `uv sync --frozen`.

## Configuration

Read once into `Settings` (pydantic-settings) in `main.py`.

| Variable | Default | Meaning |
|---|---|---|
| `DATABASE_URL` | `sqlite+aiosqlite:///./relay.db` | Selects the driver. `postgresql+asyncpg://user:pass@host/db` switches to Postgres with no code change. Plain `sqlite://` and `postgresql://` get the async driver added |
| `VERSION` | `dev` | Reported by `/health`; the Docker build arg sets it |
| `LOG_LEVEL` | `INFO` | Logs are one JSON object per line on stdout |
| `TEST_DATABASE_URL` | unset | Tests only. When set, integration tests use it instead of a temp SQLite file |

Migrations: `uv run alembic upgrade head` (reads `DATABASE_URL`).

## Docker image

```sh
docker build -t relay-backend --build-arg VERSION=dev backend
docker run -d --name relay-backend -p 8000:8000 -v relay-backend-data:/data relay-backend
curl localhost:8000/health
docker run --rm -v relay-backend-data:/data relay-backend alembic upgrade head
```

Multi-stage build: `uv sync --frozen --no-dev` in the build stage, a slim
Python runtime stage that runs as an unprivileged user, and a `HEALTHCHECK` on
`/health`. SQLite lives in the `/data` volume; set `DATABASE_URL` for Postgres.

CI: `.github/workflows/backend-test.yml` runs the gates above (pytest on
SQLite, then on a Postgres service container). `.github/workflows/backend-image.yml`
builds `linux/amd64` and `linux/arm64`, smoke-tests `/health` on the amd64
build, and publishes `ghcr.io/<owner>/<repo>-backend` (`latest`/`main`/`sha-*`
on main, `vX.Y.Z` on tags, build-only on pull requests).
