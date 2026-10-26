from typing import Protocol


class UseCase[Command, Result](Protocol):
    """A use case is a callable object built with its collaborators."""

    async def __call__(self, cmd: Command) -> Result: ...
