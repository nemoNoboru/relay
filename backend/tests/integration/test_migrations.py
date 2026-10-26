from sqlalchemy import text
from sqlalchemy.ext.asyncio import AsyncEngine

from relay.shared.infrastructure.database import create_session_factory
from relay.shared.infrastructure.unit_of_work import SqlAlchemyUnitOfWork


async def test_upgrade_head_records_revision(migrated_engine: AsyncEngine) -> None:
    async with migrated_engine.connect() as conn:
        version = (await conn.execute(text("SELECT version_num FROM alembic_version"))).scalar_one()
    assert version == "0001"


async def test_unit_of_work_runs_a_query(migrated_engine: AsyncEngine) -> None:
    uow = SqlAlchemyUnitOfWork(create_session_factory(migrated_engine))
    async with uow:
        assert (await uow.session.execute(text("SELECT 1"))).scalar_one() == 1
        await uow.commit()
