from pydantic import Field, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict

_DRIVER_PREFIXES = {
    "sqlite://": "sqlite+aiosqlite://",
    "postgresql://": "postgresql+asyncpg://",
    "postgres://": "postgresql+asyncpg://",
}


class Settings(BaseSettings):
    """Process configuration. Built once in main.py and passed to whoever needs it."""

    model_config = SettingsConfigDict(extra="ignore", frozen=True)

    database_url: str = Field(default="sqlite+aiosqlite:///./relay.db")
    version: str = Field(default="dev")
    log_level: str = Field(default="INFO")

    @field_validator("database_url")
    @classmethod
    def _use_async_driver(cls, value: str) -> str:
        """Accept plain sqlite:// and postgresql:// URLs and pick the async driver."""
        for prefix, replacement in _DRIVER_PREFIXES.items():
            if value.startswith(prefix):
                return replacement + value.removeprefix(prefix)
        return value

    @field_validator("log_level")
    @classmethod
    def _normalise_level(cls, value: str) -> str:
        level = value.upper()
        if level not in {"CRITICAL", "ERROR", "WARNING", "INFO", "DEBUG"}:
            raise ValueError(f"unknown log level: {value}")
        return level
