from typing import Any
from uuid import UUID, uuid4

from pydantic import BaseModel, ConfigDict, Field


class ValueObject(BaseModel):
    """Immutable, compared by value."""

    model_config = ConfigDict(frozen=True)


class Entity(BaseModel):
    """Has an identity; equality is by id, not by attributes."""

    model_config = ConfigDict(validate_assignment=True)

    id: UUID = Field(default_factory=uuid4)

    def __eq__(self, other: Any) -> bool:
        return isinstance(other, type(self)) and other.id == self.id

    def __hash__(self) -> int:
        return hash((type(self), self.id))
