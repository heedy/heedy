function analyze(c, vis) {
    if (c.data.size > 6 || !c.data.every((ds) => ds.length < 50000 && ds.length > 0)) {
        return vis; // Don't display table for huge datasets.
    }

    const datasets = c.data.map((data, i) => {
        // Add the timestamp and duration columns if relevant
        const columns = [{
            prop: "t",
            name: "Timestamp",
            size: 200,
            type: "timestamp",
        }];
        if (data.dt.nonNull() > 0) {
            columns.push({
                prop: "dt",
                name: "Duration",
                type: "duration",
            });
        }

        const dtype = data.d.type();
        if (dtype === "object") {
            const keys = data.d.keys();
            Object.keys(keys).forEach((key) => {
                columns.push({
                    prop: "d." + key,
                    name: key.charAt(0).toUpperCase() + key.substring(1),
                    type: data.d(key).type(),
                });
            });
        } else {
            // If the data is not objects with columns per key, just display the raw data
            columns.push({ prop: "d", name: "Data", type: dtype });
        }



        return {
            columns,
            label: c.keys[i], 
            data: c.tpl(`data[${i}]`),
            timeseries: c.tpl(`query[${i}].timeseries`),
            editable: c.tpl(`query[${i}].isSimple()`)
        };
    });

    vis.set("datatable",{
        weight: 20,
        title: "Data Table",
        type: "datatable",
        config: datasets,
    });

    return vis;
}

export default analyze;