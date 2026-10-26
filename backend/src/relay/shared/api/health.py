from typing import Annotated

from fastapi import APIRouter, Depends
from pydantic import BaseModel

from relay.container import Container
from relay.shared.api.dependencies import get_container
from relay.shared.application.health import GetHealth
from relay.shared.domain import HealthStatus

router = APIRouter()


class HealthResponse(BaseModel):
    status: str
    version: str

    @classmethod
    def from_domain(cls, health: HealthStatus) -> "HealthResponse":
        return cls(status=health.status, version=health.version)


def get_health_service(container: Annotated[Container, Depends(get_container)]) -> GetHealth:
    return container.get_health()


@router.get("/health")
async def health(service: Annotated[GetHealth, Depends(get_health_service)]) -> HealthResponse:
    return HealthResponse.from_domain(await service.execute())
