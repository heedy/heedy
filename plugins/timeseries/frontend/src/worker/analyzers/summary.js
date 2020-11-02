
import query, { dq, dtq } from "../../analysis.mjs";

function getTypeDisplay(t, series, kv) {
  let data = [
    { name: "type", value: { series: series, q: kv, transform: "type" } },
    {
      name: "length",
      value: { series: series, q: kv, transform: "length" },
    },
  ];
  switch (t) {
    case "number":
      data.push({
        name: "mean",
        value: { series: series, q: kv, transform: "mean" },
      });
      data.push({
        name: "min",
        value: { series: series, q: kv, transform: "min" },
      });
      data.push({
        name: "max",
        value: { series: series, q: kv, transform: "max" },
      });
      data.push({
        name: "stdev",
        value: { series: series, q: kv, transform: "stdev" },
      });
  }
  return data;
}

function analyze(qd) {
  if (!qd.dataset.every((ds) => ds.length >= 5)) {
    return {}; // Only display if the summary view would actually be useful...
  }
  // If it is an object dataset, extract the keys and use those
  // as tabs

  let tables = [];
  if (qd.dataset.length == 1 && dq.dataType(qd.dataset[0]) === "object") {
    let d = qd.dataset[0];
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
      data: getTypeDisplay(query(["d", kv]).dataType(d), 0, ["d", kv]),
    }));
  } else {
    tables = qd.dataset.map((d, i) => {
      let dd = getTypeDisplay(dq.dataType(d), i, ["d"]);
      let sumDuration = dtq.sum(d);
      if (sumDuration > 0) {
        dd.push({
          name: "duration",
          value: { series: i, q: ["dt"], transform: "duration" }
        });
      }
      return {
        label: `Series ${i + 1}`,
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
