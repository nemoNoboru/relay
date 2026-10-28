from relay.shared.domain import HealthStatus


class GetHealth:
    def __init__(self, version: str) -> None:
        self._version = version

    async def execute(self) -> HealthStatus:
        return HealthStatus(status="ok", version=self._version)
