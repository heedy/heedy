function extractor(q) {
  if (q.key !== undefined && q.key !== "") {
    return (dp) => (dp.d[q.key] !== undefined ? dp.d[q.key] : null);
  }
  return (dp) => dp.d;
}

function extract(ds, q, f) {
  let e = extractor(q);

  ds[q.series].forEach((dp) => {
    let v = e(dp);
    if (v !== null) {
      f(v);
    }
  });
}

let transforms = {
  length(ds, q) {
    if (q.key !== undefined && q.key !== "") {
      // We need to actually count the non-null values
      let count = 0;
      ds[q.series].forEach((dp) => {
        if (dp.d[q.key] !== undefined && dp.d[q.key] !== null) {
          count++;
        }
      });
      return count;
    }

    return ds[q.series].length;
  },
  type(ds, q) {
    if (q.key !== undefined && q.key !== "") {
      return ds[q.series].keyType(q.key);
    }
    return ds[q.series].dataType();
  },
  sum(ds, q) {
    let count = 0;
    extract(ds, q, (v) => {
      count += v;
    });

    return count;
  },
  mean(ds, q) {
    return transforms.sum(ds, q) / transforms.length(ds, q);
  },
  min(ds, q) {
    let curval = Infinity;
    extract(ds, q, (v) => {
      if (v < curval) {
        curval = v;
      }
    });
    return curval;
  },
  max(ds, q) {
    let curval = -Infinity;
    extract(ds, q, (v) => {
      if (v > curval) {
        curval = v;
      }
    });
    return curval;
  },
  stdev(ds, q) {
    let curval = 0;
    let mean = transforms.mean(ds, q);
    let length = transforms.length(ds, q);
    extract(ds, q, (v) => {
      curval += Math.pow(v - mean, 2);
    });
    return Math.sqrt(curval / (length - 1));
  },
};
function getData(qd, query) {
  if (typeof query !== "object") {
    return query; // Objects are considered queries
  }

  if (transforms[query.transform] !== undefined) {
    return transforms[query.transform](qd.dataset, query);
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
