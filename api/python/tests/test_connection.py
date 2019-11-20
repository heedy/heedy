import pytest
from heedy import App


def test_basics():
    c = App("gV9A4/J0eS4nLGAnzYyn")
    assert len(c.listObjects()) > 0


@pytest.mark.asyncio
async def test_basics_async():
    c = App("gV9A4/J0eS4nLGAnzYyn", session="async")
    assert len(await c.listObjects()) > 0
    print(await c.read())
