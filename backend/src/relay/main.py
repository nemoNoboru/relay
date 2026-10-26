import logging
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager

from fastapi import FastAPI

from relay.container import Container
from relay.shared.api import health
from relay.shared.config import Settings
from relay.shared.infrastructure.database import create_engine, create_session_factory
from relay.shared.logging import configure_logging

logger = logging.getLogger(__name__)


def create_app(settings: Settings) -> FastAPI:
    """Composition root: the only place that builds Settings-dependent objects."""

    @asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        configure_logging(settings.log_level)
        engine = create_engine(settings.database_url)
        app.state.container = Container(settings, create_session_factory(engine))
        logger.info("startup", extra={"version": settings.version})
        try:
            yield
        finally:
            await engine.dispose()
            logger.info("shutdown")

    app = FastAPI(title="Relay", version=settings.version, lifespan=lifespan)
    app.include_router(health.router)
    return app


app = create_app(Settings())
