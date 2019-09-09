import pytest
from heedy import Connection


def test_basics():
    c = Connection("gV9A4/J0eS4nLGAnzYyn")
    assert len(c.listSources()) > 0

@pytest.mark.asyncio
async def test_basics_async():
    c = Connection("gV9A4/J0eS4nLGAnzYyn",session="async")
    assert len(await c.listSources()) > 0
    print(await c.read())