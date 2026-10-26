"""Domain building blocks: pure Pydantic models, no framework imports."""

from relay.shared.domain.errors import ConflictError, DomainError, NotFoundError, ValidationError
from relay.shared.domain.events import DomainEvent
from relay.shared.domain.health import HealthStatus
from relay.shared.domain.models import Entity, ValueObject
from relay.shared.domain.repository import Repository

__all__ = [
    "ConflictError",
    "DomainError",
    "DomainEvent",
    "Entity",
    "HealthStatus",
    "NotFoundError",
    "Repository",
    "ValidationError",
    "ValueObject",
]
