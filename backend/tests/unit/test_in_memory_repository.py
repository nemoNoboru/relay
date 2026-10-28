from uuid import uuid4

from relay.shared.domain import Entity
from tests.fakes import InMemoryRepository


class Note(Entity):
    body: str


async def test_returns_what_was_added() -> None:
    repo = InMemoryRepository[Note]()
    note = Note(body="hi")
    await repo.add(note)
    assert await repo.get(note.id) == note


async def test_returns_none_for_unknown_id() -> None:
    assert await InMemoryRepository[Note]().get(uuid4()) is None
