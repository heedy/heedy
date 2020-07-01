import pytest
from heedy import App


def lists_equal(l1, l2):
    assert len(l1) == len(l2)
    for i in range(len(l1)):
        assert l1[i] == l2[i]


def test_dashboard():
    a = App("testkey")
    t1 = a.objects.create("obj1", key="key1")
    t2 = a.objects.create("obj2", key="key2")
    t3 = a.objects.create("obj3", key="key3")

    t1.insert_array(
        [{"t": 1, "d": 1}, {"t": 2, "d": 1},]
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
    d = a.objects.create("db", otype="dashboard")

    d.add(
        {
            "type": "dataset",
            "query": [
                {
                    "merge": [
                        {"timeseries": t1.id},
                        {"timeseries": t2.id, "t1": 3.0},
                        {"timeseries": t3.id, "i1": 1, "i2": 2},
                    ]
                }
            ],
            "events": [{"event": "timeseries_data_write", "object_id": t1.id}],
        }
    )

    elements = d[:]
    assert len(elements) == 1

    lists_equal(
        elements[0].data[0],
        [
            {"t": 1, "d": 1},
            {"t": 2, "d": 1},
            {"t": 2.2, "d": 3},
            {"t": 3.1, "d": 2},
            {"t": 4.1, "d": 2},
            {"t": 5.1, "d": 2},
        ],
    )

    t1.insert_array(
        [{"t": 3, "d": 1}, {"t": 4, "d": 1}, {"t": 5, "d": 1},]
    )

    element = d[elements[0].id]

    lists_equal(
        element.data[0],
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

