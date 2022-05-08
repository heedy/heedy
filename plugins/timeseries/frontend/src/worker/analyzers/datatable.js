import query, { dq } from "../../analysis.mjs";

function canEdit(q) {
  if (q.transform !== undefined && q.transform.length >0) {
    return false;
  }
  if (q.post_transform !== undefined && q.post_transform.length >0) {
    return false;
  }
  if (q.timeseries === undefined) {
    return false;
  }
  if (q.dataset!==undefined) {
    return false;
  }
  return true;
}

function analyze(qd,vis) {
  if (
    qd.keys.length > 6 ||
    !qd.dataset_array.every((ds) => ds.length < 50000 && ds.length > 0)
  ) {
    return vis; // Don't display table for huge datasets.
  }
  const datasets = qd.dataset_array.map((data, i) => {

    // Add the timestamp and duration columns if relevant
    let columns = [{ prop: "t", name: "Timestamp", size: 200, type: "timestamp" }];
    if (data.some((dp) => dp.dt !== undefined)) {
      columns.push({
        prop: "dt",
        name: "Duration",
        type: "duration",
      });
    }

    const dtype = dq.dataType(data);
    if (dtype === "object") {
      const keys = dq.keys(data);
      Object.keys(keys).forEach((key) => {
        columns.push({
          prop: "d." + key,
          name: key.charAt(0).toUpperCase() + key.substring(1),
          type: query(["d", key]).dataType(data),
        });
      });

    } else {
      // If the data is not objects with columns per key, just display the raw data
      columns.push({ prop: "d", name: "Data", type: dtype });
    }

    // Now determine whether the data can be edited. This is only
    // possible if the data is from a single timeseries, and does not have any transforms active.

    const editable = canEdit(qd.query[qd.keys[i]]);
    let timeseries = null;
    if (editable) {
      timeseries = qd.query[qd.keys[i]].timeseries;
    }

    return { columns, label: qd.keys[i], editable: editable,timeseries: timeseries };
  });

  vis.datatable = {
    weight: 20,
    title: "Data Table",
    visualization: "datatable",
    config: datasets,
  };

  return vis;
}

export default analyze;
