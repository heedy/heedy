import pytest
from heedy import App

"""Dashboard has been temporarily disabled
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
        [
            {
                "merge": [
                    {"timeseries": t1.id},
                    {"timeseries": t2.id, "t1": 3.0},
                    {"timeseries": t3.id, "i1": 1, "i2": 2},
                ]
            }
        ]
    )

    elements = d[:]
    assert len(elements) == 1

    assert elements[0].title == ""
    elements[0].title = "Hello World"
    assert elements[0].title == "Hello World"
    assert elements[0].settings == {}
    elements[0].settings = {"hi": "ho"}
    assert elements[0].settings == {"hi": "ho"}

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


def test_dashboard_indices():
    a = App("testkey")
    t = a.objects.create("obj1", key="key1")
    t.insert_array(
        [{"t": 1, "d": 1}, {"t": 2, "d": 1},]
    )
    d = a.objects.create("db", otype="dashboard")

    d.add([{"timeseries": t.id}], title="lol")
    d.add([{"timeseries": t.id}], title="lol2")
    d.add([{"timeseries": t.id}], title="lol3")
    d.add([{"timeseries": t.id}], title="lol4")

    els = d[:]
    assert len(els) == 4
    assert els[0].title == "lol"
    assert els[1]["title"] == "lol2"
    assert els[2]["title"] == "lol3"
    assert els[3]["title"] == "lol4"

    els[2].index = 1  # switch lol2 and lol3

    els = d[:]
    assert len(els) == 4
    assert els[0].title == "lol" and els[0].index == 0
    assert els[1]["title"] == "lol3" and els[1]["index"] == 1
    assert els[2]["title"] == "lol2" and els[2]["index"] == 2
    assert els[3]["title"] == "lol4" and els[3]["index"] == 3

    els[1].index = 2  # switch back

    els = d[:]
    assert len(els) == 4
    assert els[0].title == "lol" and els[0].index == 0
    assert els[1]["title"] == "lol2" and els[1]["index"] == 1
    assert els[2]["title"] == "lol3" and els[2]["index"] == 2
    assert els[3]["title"] == "lol4" and els[3]["index"] == 3

    els[1].delete()

    els = d[:]
    assert len(els) == 3
    assert els[0].title == "lol" and els[0].index == 0
    assert els[1]["title"] == "lol3" and els[1]["index"] == 1
    assert els[2]["title"] == "lol4" and els[2]["index"] == 2

    els[2].delete()

    els = d[:]
    assert len(els) == 2
    assert els[0].title == "lol" and els[0].index == 0
    assert els[1]["title"] == "lol3" and els[1]["index"] == 1

    els[0].delete()
    els[1].delete()

    els = d[:]
    assert len(els) == 0

    for o in a.objects():
        o.delete()
"""