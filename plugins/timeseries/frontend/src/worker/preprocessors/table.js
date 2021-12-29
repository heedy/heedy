import moment from "../../../dist/moment.mjs";

import query from "../../analysis.mjs";

let transforms = {
  count(ds, q) {
    if (q.q.length > 1) {
      // We need to actually count the non-null values
      return query(q.q).nonNull(ds[q.key]);
    }
    return ds[q.key].length;
  },
  type(ds, q) {
    return query(q.q).dataType(ds[q.key]);
  },
  sum(ds, q) {
    return query(q.q).sum(ds[q.key]);
  },
  duration(ds, q) {
    return moment.duration(query(q.q).sum(ds[q.key]), "seconds").humanize();
  },
  mean(ds, q) {
    return query(q.q).mean(ds[q.key]);
  },
  min(ds, q) {
    return query(q.q).min(ds[q.key]);
  },
  max(ds, q) {
    return query(q.q).max(ds[q.key]);
  },
  stdev(ds, q) {
    return query(q.q).stddev(ds[q.key]);
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
