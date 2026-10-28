from relay.shared.application.health import GetHealth


async def test_reports_ok_and_version() -> None:
    result = await GetHealth(version="9.9.9").execute()
    assert result.status == "ok"
    assert result.version == "9.9.9"
