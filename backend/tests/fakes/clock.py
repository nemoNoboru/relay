from datetime import UTC, datetime

from relay.shared.application import Clock


class FixedClock(Clock):
    def __init__(self, now: datetime | None = None) -> None:
        self._now = now or datetime(2026, 1, 1, tzinfo=UTC)

    def now(self) -> datetime:
        return self._now
