class DomainError(Exception):
    """Base class for errors raised by domain and application code."""


class NotFoundError(DomainError):
    """The requested entity does not exist."""


class ConflictError(DomainError):
    """The operation conflicts with the current state."""


class ValidationError(DomainError):
    """A business rule rejected the input."""
