from types import TracebackType
from typing import Self

from relay.shared.application import UnitOfWork


class FakeUnitOfWork(UnitOfWork):
    """In-memory stand-in for a UnitOfWork. Contexts add InMemory<X>Repository attributes."""

    def __init__(self) -> None:
        self.committed = False
        self.rolled_back = False

    async def __aenter__(self) -> Self:
        return self

    async def __aexit__(
        self,
        exc_type: type[BaseException] | None,
        exc: BaseException | None,
        tb: TracebackType | None,
    ) -> None:
        if exc_type is not None:
            await self.rollback()

    async def commit(self) -> None:
        self.committed = True

    async def rollback(self) -> None:
        self.rolled_back = True
