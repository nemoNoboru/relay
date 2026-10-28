from collections.abc import AsyncIterator
from pathlib import Path

import pytest
from httpx import ASGITransport, AsyncClient

from relay.main import create_app
from relay.shared.config import Settings


@pytest.fixture
def dist(tmp_path: Path) -> Path:
    root = tmp_path / "dist"
    (root / "_expo").mkdir(parents=True)
    (root / "index.html").write_text("<p>index</p>")
    (root / "login.html").write_text("<p>login</p>")
    (root / "_expo" / "app.js").write_text("console.log(1)")
    (tmp_path / "secret.txt").write_text("secret")
    return root


@pytest.fixture
async def web_client(test_settings: Settings, dist: Path) -> AsyncIterator[AsyncClient]:
    app = create_app(test_settings.model_copy(update={"web_dist_dir": dist}))
    async with (
        app.router.lifespan_context(app),
        AsyncClient(transport=ASGITransport(app=app), base_url="http://test") as http,
    ):
        yield http


async def test_serves_index_at_root(web_client: AsyncClient) -> None:
    response = await web_client.get("/")
    assert response.status_code == 200
    assert response.text == "<p>index</p>"


async def test_serves_static_assets(web_client: AsyncClient) -> None:
    response = await web_client.get("/_expo/app.js")
    assert response.text == "console.log(1)"


async def test_route_resolves_to_its_exported_html(web_client: AsyncClient) -> None:
    assert (await web_client.get("/login")).text == "<p>login</p>"


async def test_unknown_route_falls_back_to_index(web_client: AsyncClient) -> None:
    response = await web_client.get("/some/deep/link")
    assert response.status_code == 200
    assert response.text == "<p>index</p>"


async def test_api_still_wins_over_the_web_app(web_client: AsyncClient) -> None:
    assert (await web_client.get("/api/health")).json()["status"] == "ok"
    assert (await web_client.get("/health")).json()["status"] == "ok"


async def test_unknown_api_path_is_404_not_the_web_app(web_client: AsyncClient) -> None:
    assert (await web_client.get("/api/nope")).status_code == 404


async def test_cannot_escape_the_dist_dir(web_client: AsyncClient) -> None:
    response = await web_client.get("/%2e%2e/secret.txt")
    assert "secret" not in response.text


async def test_web_is_off_when_unset(client: AsyncClient) -> None:
    assert (await client.get("/")).status_code == 404
