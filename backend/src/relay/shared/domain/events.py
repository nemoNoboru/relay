from datetime import UTC, datetime

from pydantic import BaseModel, ConfigDict, Field


class DomainEvent(BaseModel):
    """Something that happened in the domain. Past-tense subclasses add the payload."""

    model_config = ConfigDict(frozen=True)

    occurred_at: datetime = Field(default_factory=lambda: datetime.now(UTC))
