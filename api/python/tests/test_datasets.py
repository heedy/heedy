import pytest
from heedy import App
from heedy.datasets import Merge, Dataset


def lists_equal(l1, l2):
    assert len(l1) == len(l2)
    for i in range(len(l1)):
        assert l1[i] == l2[i]


def test_merge():
    a = App("testkey")
    t1 = a.objects.create("obj1", key="key1")
    t2 = a.objects.create("obj2", key="key2")
    t3 = a.objects.create("obj3", key="key3")

    t1.insert_array(
        [
            {"t": 1, "d": 1},
            {"t": 2, "d": 1},
            {"t": 3, "d": 1},
            {"t": 4, "d": 1},
            {"t": 5, "d": 1},
        ]
    )
    t2.insert_array(
        [
            {"t": 1.1, "d": 2},
            {"t": 2.1, "d": 2},
            {"t": 3.1, "d": 2},
            {"t": 4.1, "d": 2},
            {"t": 5.1, "d": 2},
        ]
    )
    t3.insert_array(
        [
            {"t": 1.2, "d": 3},
            {"t": 2.2, "d": 3},
            {"t": 3.3, "d": 3},
            {"t": 4.4, "d": 3},
            {"t": 5.5, "d": 3},
        ]
    )

    m = Merge(a)
    m.add(t1)
    m.add(t2.id, t1=3.0)
    m.add(t3, i1=1, i2=2)

    result = m.run()

    lists_equal(
        result,
        [
            {"t": 1, "d": 1},
            {"t": 2, "d": 1},
            {"t": 2.2, "d": 3},
            {"t": 3, "d": 1},
            {"t": 3.1, "d": 2},
            {"t": 4, "d": 1},
            {"t": 4.1, "d": 2},
            {"t": 5, "d": 1},
            {"t": 5.1, "d": 2},
        ],
    )

    for o in a.objects():
        o.delete()


def test_tdataset():
    a = App("testkey")
    t1 = a.objects.create("obj1", key="key1")

    t1.insert_array(
        [{"t": 2, "d": 73}, {"t": 5, "d": 84}, {"t": 8, "d": 81}, {"t": 11, "d": 79}]
    )

    ds = Dataset(a, t1=0, t2=8.1, dt=2)

    ds.add("temperature", t1)

    res = ds.run()

    assert 5 == len(res)
    lists_equal(
        res,
        [
            {"t": 0, "d": {"temperature": 73}},
            {"t": 2, "d": {"temperature": 73}},
            {"t": 4, "d": {"temperature": 84}},
            {"t": 6, "d": {"temperature": 84}},
            {"t": 8, "d": {"temperature": 81}},
        ],
    )

    for o in a.objects():
        o.delete()
