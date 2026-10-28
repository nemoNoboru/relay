import pytest
from httpx import AsyncClient


@pytest.mark.parametrize("path", ["/api/health", "/health"])
async def test_health_returns_ok_and_version(client: AsyncClient, path: str) -> None:
    response = await client.get(path)
    assert response.status_code == 200
    assert response.json() == {"status": "ok", "version": "test"}
