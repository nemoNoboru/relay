from typing import Annotated

from fastapi import APIRouter, Depends

from relay.container import Container
from relay.shared.api.dependencies import get_container
from relay.shared.application.health import GetHealth, HealthDTO

router = APIRouter()


def get_health_use_case(container: Annotated[Container, Depends(get_container)]) -> GetHealth:
    return container.get_health()


@router.get("/health")
async def health(use_case: Annotated[GetHealth, Depends(get_health_use_case)]) -> HealthDTO:
    return await use_case()
