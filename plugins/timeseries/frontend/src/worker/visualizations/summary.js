function getSummaryForType(type,access) {
    let summary = [
        {name: "type", value: type},
        {name: "count", value: access("nonNull")}
    ];

    switch (type) {
        case "number":
            summary.push({name: "mean", value: access("mean")});
            summary.push({name: "min", value: access("min")});
            summary.push({name: "max", value: access("max")});
            summary.push({name: "stdev", value: access("stdev")});
    }

    return summary;
}

function summary(c,vis) {
    if (!c.data.every((d) => d.length >=5)) {
        return; // Only display if a summary view would actually be useful
    }
    // if it is an object, extract the keys, and use those as tabs.
    let tables = [];
    if (c.data.length == 1 && c.data[0].d.type() === "object") {
        const d = c.data[0].d;
        const k = d.keys();

        if (k["latitude"] !== undefined || k["longitude"] !== undefined) {
            return;
        }

        const karr = Object.keys(k);
        karr.sort();
        tables = karr.map(kv => ({
            label: kv,
            columns: [
                { prop: "name", name: "Quantity" },
                { prop: "value", name: "Value" },
            ],
            data: getSummaryForType(d(kv).type(),(fname)=> c.tpl(`data[0].d(${c.tpls(kv)}).${fname}()`))
        }));
    } else {
        tables = c.data.map((dpa,i) => {
            const dd = getSummaryForType(dpa.d.type(),(fname)=> c.tpl(`data[${i}].d.${fname}()`));
            if (dpa.dt.sum() > 0) {
                dd.push({name: "duration", value: c.tpl(`data[${i}].dt.sum()`),"value.type": "duration"});
            }
            return {
                label: c.getSeriesLabelTemplate(i),
                columns: [
                    { prop: "name", name: "Quantity" },
                    { prop: "value", name: "Value" },
                  ],
                data: dd,
            };
        });
    }

    vis.set("summary", {
        weight: 19,
        title: "Summary",
        type: "table",
        config: tables
    });

    return;
}

export default summary;