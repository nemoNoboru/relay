from datetime import UTC, datetime


class Clock:
    """The current time. Tests subclass it to pin the time."""

    def now(self) -> datetime:
        return datetime.now(UTC)
