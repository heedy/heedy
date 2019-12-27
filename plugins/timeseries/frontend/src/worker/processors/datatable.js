import moment from "../../../dist/moment.mjs";
async function process(object, data) {
  if (data.length == 0 || data.length > 10000) {
    return {};
  }

  if (typeof data[0].d !== "object") {
    // It is not an object, so we simply dump the data
    let datapoints = data.map(d => ({
      t: new Date(d.t * 1000).toLocaleString(),
      d: d.d.toString()
    }));

    let config = [
      { prop: "t", name: "Timestamp" },
      { prop: "d", name: "Data" }
    ];

    if (data.some(dp => dp.td !== undefined)) {
      for (let i = 0; i < data.length; i++) {
        datapoints[i].dt =
          data[i].td !== undefined
            ? moment.duration(data[i].td, "seconds").humanize()
            : "";
      }
      config = [
        { prop: "t", name: "Timestamp" },
        { prop: "dt", name: "Duration" },
        { prop: "d", name: "Data" }
      ];
    }
    return {
      datatable: {
        weight: 20,
        title: "Data Table",
        view: "datatable",
        data: {
          config: config,
          data: datapoints
        }
      }
    };
  }

  // It is an object
  // TODO
  return {};
}

export default process;
