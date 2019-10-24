async function process(source, data) {
    if (data.length == 0) {
        return {};
    }

    if (typeof data[0].d !== "object") {
        // It is not an object, so we simply dump the data
        return {
            datatable: {
                weight: 10,
                title: "Data Table",
                view: "datatable",
                data: {
                    header: ['Data'],
                    data: data.map(d => ({
                        t: d.t,
                        d: [d.d],
                        key: JSON.stringify(d)
                    }))
                }
            }
        }
    }


    // It is an object
    // TODO
    return {};
}

export default process;