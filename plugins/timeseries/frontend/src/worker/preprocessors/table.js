import moment from "../../../dist/moment.mjs";

import query from "../../analysis.mjs";

let transforms = {
  length(ds, q) {
    if (q.q.length > 1) {
      // We need to actually count the non-null values
      return query(q.q).nonNull(ds[q.series]);
    }
    return ds[q.series].length;
  },
  type(ds, q) {
    return query(q.q).dataType(ds[q.series]);
  },
  sum(ds, q) {
    return query(q.q).sum(ds[q.series]);
  },
  duration(ds, q) {
    return moment.duration(query(q.q).sum(ds[q.series]) / 1000, "seconds").humanize();
  },
  mean(ds, q) {
    return query(q.q).mean(ds[q.series]);
  },
  min(ds, q) {
    return query(q.q).min(ds[q.series]);
  },
  max(ds, q) {
    return query(q.q).max(ds[q.series]);
  },
  stdev(ds, q) {
    return query(q.q).stddev(ds[q.series]);
  },
};
function getData(qd, qq) {
  if (typeof qq !== "object") {
    return qq; // Objects are considered queries
  }

  if (transforms[qq.transform] !== undefined) {
    return transforms[qq.transform](qd.dataset, qq);
  }
  return "?";
}

function preprocessor(qd, visualization) {
  return {
    ...visualization,
    type: "table",
    config: visualization.config.map((c) => ({
      ...c,
      data: c.data.map((d) =>
        Object.keys(d).reduce((o, k) => {
          let newo = { ...o };
          newo[k] = getData(qd, d[k]);
          return newo;
        }, {})
      ),
    })),
  };
}

export default preprocessor;
