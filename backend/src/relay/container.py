from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker

from relay.shared.application import Clock, UnitOfWorkFactory
from relay.shared.application.health import GetHealth
from relay.shared.config import Settings
from relay.shared.infrastructure.clock import SystemClock
from relay.shared.infrastructure.unit_of_work import SqlAlchemyUnitOfWork


class Container:
    """Constructor wiring. Holds long-lived objects and builds use cases on demand.

    Later contexts add one `get_<use_case>()` method each; the api layer's
    providers call them. Nothing here is a global.
    """

    def __init__(
        self,
        settings: Settings,
        session_factory: async_sessionmaker[AsyncSession],
        clock: Clock | None = None,
    ) -> None:
        self.settings = settings
        self.session_factory = session_factory
        self.clock: Clock = clock or SystemClock()

    def uow_factory(self) -> UnitOfWorkFactory:
        return lambda: SqlAlchemyUnitOfWork(self.session_factory)

    def get_health(self) -> GetHealth:
        return GetHealth(version=self.settings.version)
