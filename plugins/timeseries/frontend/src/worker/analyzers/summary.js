function getTypeDisplay(t, series, kv) {
  let data = [
    { name: "type", value: { series: series, key: kv, transform: "type" } },
    {
      name: "length",
      value: { series: series, key: kv, transform: "length" },
    },
  ];
  switch (t) {
    case "number":
      data.push({
        name: "mean",
        value: { series: series, key: kv, transform: "mean" },
      });
      data.push({
        name: "min",
        value: { series: series, key: kv, transform: "min" },
      });
      data.push({
        name: "max",
        value: { series: series, key: kv, transform: "max" },
      });
      data.push({
        name: "stdev",
        value: { series: series, key: kv, transform: "stdev" },
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
  if (qd.dataset.length == 1 && qd.dataset[0].dataType() === "object") {
    let d = qd.dataset[0];
    let k = d.keys();

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
      data: getTypeDisplay(d.keyType(kv), 0, kv),
    }));
  } else {
    tables = qd.dataset.map((d, i) => {
      return {
        label: `Series ${i + 1}`,
        columns: [
          { prop: "name", name: "Quantity" },
          { prop: "value", name: "Value" },
        ],
        data: getTypeDisplay(d.dataType(), i, ""),
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
