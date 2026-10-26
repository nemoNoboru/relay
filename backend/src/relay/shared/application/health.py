from pydantic import BaseModel


class HealthDTO(BaseModel):
    status: str
    version: str


class GetHealth:
    def __init__(self, version: str) -> None:
        self._version = version

    async def __call__(self) -> HealthDTO:
        return HealthDTO(status="ok", version=self._version)
