async function analyze(qd) {
  let cols = qd.dataset.map((data) => {
    let columns = [{ prop: "t", name: "Timestamp" }];
    if (data.some((dp) => dp.dt !== undefined)) {
      columns.push({
        prop: "dt",
        name: "Duration",
      });
    }

    if (typeof data[0].d !== "object") {
      columns.push({ prop: "d", name: "Data" });
    } else {
      // It is an object, so find the properties, and make them table headers rather than just the raw data
      let headers = {};
      let isWeird = false;
      data.forEach((dp) => {
        if (typeof dp.d !== "object") {
          isWeird = true;
        } else {
          Object.keys(dp.d).forEach((k) => {
            headers[k] = true;
          });
        }
      });

      if (isWeird) {
        // Just give the raw data, since wtf
        columns.push({ prop: "d", name: "Data" });
      } else {
        Object.keys(headers).forEach((k) => {
          columns.push({ prop: k, name: k });
        });
      }
    }
    return { columns };
  });

  return {
    datatable: {
      weight: 20,
      title: "Data Table",
      visualization: "datatable",
      config: cols,
    },
  };
}

export default analyze;
