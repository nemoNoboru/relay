from uuid import UUID

from relay.shared.domain import Entity, Repository


class InMemoryRepository[E: Entity](Repository[E, UUID]):
    """Dict-backed repository. Context fakes subclass it and add their own queries."""

    def __init__(self) -> None:
        self._items: dict[UUID, E] = {}

    async def get(self, id: UUID) -> E | None:
        return self._items.get(id)

    async def add(self, entity: E) -> None:
        self._items[entity.id] = entity
