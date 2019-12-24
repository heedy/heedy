async function process(object, data) {
  if (data.length == 0) {
    return {};
  }
  if (data.length > 1000) {
    return {}; // TEMPORARY, while we don't have a virtual table
  }

  if (typeof data[0].d !== "object") {
    // It is not an object, so we simply dump the data
    return {
      datatable: {
        weight: 20,
        title: "Data Table",
        view: "datatable",
        data: {
          header: ["Data"],
          data: data.map(d => ({
            t: d.t,
            d: [d.d],
            key: JSON.stringify(d)
          }))
        }
      }
    };
  }

  // It is an object
  // TODO
  return {};
}

export default process;
