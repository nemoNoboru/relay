import pytest

from tests.fakes import FakeUnitOfWork


async def test_rolls_back_when_the_block_raises() -> None:
    uow = FakeUnitOfWork()
    with pytest.raises(RuntimeError):
        async with uow:
            raise RuntimeError("boom")
    assert uow.rolled_back
    assert not uow.committed
