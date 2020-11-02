
import { getType, getKeys, getMin, getMax, getSum, getVar, getNonNull } from "./analysis/util.js";


let qprops = {
  "keys": getKeys,
  "dataType": getType,
  "min": getMin,
  "max": getMax,
  "sum": getSum,
  "nonNull": getNonNull,
  "stddev": (f, ts) => Math.sqrt((getVar(f, ts) - Math.pow(f.mean(ts), 2)) / (f.nonNull(ts) - 1))
};



function addProps(f) {
  f.cache = new WeakMap();

  // get is the underlying cache handler function,
  // which allows caching results of computation for many
  // analysis functions.
  f.get = (key, ff, arr) => {
    let cval = {};
    if (f.cache.has(arr)) {
      cval = f.cache.get(arr);
      if (cval[key] !== undefined) {
        return cval[key];
      }
    }
    let res = ff(f, arr);
    cval[key] = res;
    f.cache.set(arr, cval);

    return res;
  }

  f.set = (key, ff) => {
    f[key] = (arr) => f.get(key, ff, arr);
  }

  // Set up the properties
  Object.keys(qprops).forEach(k => f.set(k, qprops[k]));

  f.isNumeric = (arr) => {
    let t = f.dataType(arr)
    return t === "number" || t === "boolean";
  }
  f.isBoolean = (arr) => f.dataType(arr) === "boolean";

  f.mean = (arr) => f.sum(arr) / f.nonNull(arr);

  f.noNulls = (arr) => f.nonNull(arr) == arr.length;

  return f;
}


function generateQuery(query) {
  let f = (dp) => {
    for (let i = 0; i < query.length; i++) {
      dp = dp[query[i]];
      if (dp === undefined) {
        return null;
      }
    }
    return dp;
  };
  addProps(f);
  return f;
}



let dq = (dp) => dp.d;
let tq = (dp) => dp.t * 1000; // Timestamps are multiplied by 1000
let dtq = (dp) => (dp.dt === undefined ? 0 : dp.dt * 1000);

addProps(dq);
addProps(tq);
addProps(dtq);

let qcache = {};

function query(q) {
  if (q.length == 1) {
    switch (q[0]) {
      case "d":
        return dq;
      case "t":
        return tq;
      case "dt":
        return dtq;
    }
  }
  let key = JSON.stringify(q.slice(1));
  if (qcache[key] === undefined) {
    qcache[key] = generateQuery(q);
  }
  return qcache[key];
}


// setQueryProp allows external stuff to set properties of a query that will cache results
function setQueryProp(key, ff) {
  qprops[key] = ff;

  // Also add the prop to the existing queries
  dq.set(key, ff);
  tq.set(key, ff);
  dtq.set(key, ff);
  Object.keys(qcache, k => qcache[k].set(key, ff));
}

export {
  dq, tq, dtq, getKeys, getType, setQueryProp
}

export default query;