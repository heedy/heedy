
import query, { dq, dtq } from "../../analysis.mjs";

function getTypeDisplay(t, key, kv) {
  let data = [
    { name: "type", value: { key: key, q: kv, transform: "type" } },
    {
      name: "count",
      value: { key: key, q: kv, transform: "count" },
    },
  ];
  switch (t) {
    case "number":
      data.push({
        name: "mean",
        value: { key: key, q: kv, transform: "mean" },
      });
      data.push({
        name: "min",
        value: { key: key, q: kv, transform: "min" },
      });
      data.push({
        name: "max",
        value: { key: key, q: kv, transform: "max" },
      });
      data.push({
        name: "stdev",
        value: { key: key, q: kv, transform: "stdev" },
      });
  }
  return data;
}

function analyze(qd) {
  if (!qd.dataset_array.every((ds) => ds.length >= 5)) {
    return {}; // Only display if the summary view would actually be useful...
  }
  // If it is an object dataset, extract the keys and use those
  // as tabs

  let tables = [];
  if (qd.dataset_array.length == 1 && dq.dataType(qd.dataset_array[0]) === "object") {
    let d = qd.dataset_array[0];
    let k = dq.keys(d);

    if (k["latitude"] !== undefined || k["longitude"] !== undefined) {
      return {};
    }

    let karr = Object.keys(k);
    karr.sort();

    tables = karr.map((kv) => ({
      label: kv,
      columns: [
        { prop: "name", name: "Quantity" },
        { prop: "value", name: "Value" },
      ],
      data: getTypeDisplay(query(["d", kv]).dataType(d), qd.keys[0], ["d", kv]),
    }));
  } else {
    tables = qd.dataset_array.map((d, i) => {
      let dd = getTypeDisplay(dq.dataType(d), qd.keys[i], ["d"]);
      let sumDuration = dtq.sum(d);
      if (sumDuration > 0) {
        dd.push({
          name: "duration",
          value: { key: qd.keys[i], q: ["dt"], transform: "duration" }
        });
      }
      return {
        label: qd.keys[i],
        columns: [
          { prop: "name", name: "Quantity" },
          { prop: "value", name: "Value" },
        ],
        data: dd,
      };
    });
  }

  return {
    summary: {
      weight: 19,
      title: "Summary",
      visualization: "table",
      config: tables,
    },
  };
}

export default analyze;
