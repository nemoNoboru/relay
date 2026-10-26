from datetime import datetime
from types import TracebackType
from typing import Protocol, Self


class Clock(Protocol):
    def now(self) -> datetime: ...


class UnitOfWork(Protocol):
    """A transaction. Contexts extend this Protocol with their repositories."""

    async def __aenter__(self) -> Self: ...

    async def __aexit__(
        self,
        exc_type: type[BaseException] | None,
        exc: BaseException | None,
        tb: TracebackType | None,
    ) -> None: ...

    async def commit(self) -> None: ...

    async def rollback(self) -> None: ...


class UnitOfWorkFactory(Protocol):
    def __call__(self) -> UnitOfWork: ...
