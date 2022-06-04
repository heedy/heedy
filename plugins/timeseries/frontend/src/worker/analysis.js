function cleanDT(ts) {
    // TODO: This is actually a bug in the underlying heedy code when merging timeseries with durations
    for (let i = 0; i < ts.length - 1; i++) {
        if (ts[i].dt !== undefined && ts[i].t + ts[i].dt > ts[i + 1].t) {
            ts[i].dt = ts[i + 1].t - ts[i].t;
        }
    }
    return ts;
}

function flatten(ts, depth = 1) {
    return ts.map((dp) => {
        let res = { ...dp };
        Object.keys(dp.d).forEach((k) => {
            res["d." + k] = dp.d[k];
        });
        return res;
    });
}

/**
 * Splits datapoints with durations into two elements - one at start of the duration,
 * and one at end of the duration
 *
 * @param {*} ts timeseries
 */
function explicitDuration(ts, offset = 0.001) {
    let res = new Array(ts.length * 2);
    let j = 0;
    for (let i = 0; i < ts.length; i++) {
        res[j] = ts[i];
        j++;
        if (ts[i].dt !== undefined && ts[i].dt != 0) {
            res[j] = {
                t: ts[i].t + ts[i].dt - offset,
                d: ts[i].d,
            };
            j++;
        }
    }
    return res.slice(0, j);
}




function getType(p,ts,extractor) {
    if (ts.length == 0) {
        return "";
    }
    let i = 0;
    for (; i < ts.length; i++) {
        if (extractor(ts[0]) !== null) {
            break;
        }
    }
    if (i == ts.length) {
        return ""; // All null
    }
    let curtype = typeof extractor(ts[i]);
    if (
        ts.every((dp) => typeof extractor(dp) === curtype || extractor(dp) === null)
    ) {
        if (curtype != "string") {
            return curtype;
        }
        // Check if the data is categorical (enum) if it is a string
        let vals = {};
        let uniques = 0;
        for (i = 0; i < ts.length; i++) {
            let v = extractor(ts[i]);
            if (vals[v] === undefined) {
                vals[v] = 1;
                uniques++;
                if (uniques > 100 || uniques > ts.length / 3) {
                    return "string";
                }
            }
        }
        return "enum";
    }
    return "";
}

function getKeys(p,ts,f) {
    let vals = {};
    ts.forEach((dp) => {
        Object.keys(f(dp)).forEach((k) => {
            if (vals[k] === undefined) {
                vals[k] = 0;
            }
            vals[k]++;
        });
    });
    return vals;
}


function getMin(p,ts,f) {
    return ts.reduce((cur, dp) => {
        let v = f(dp);
        if (v == null || v >= cur) {
            return cur;
        }
        return v;
    }, Infinity)
}

function getMax(p,ts,f) {
    return ts.reduce((cur, dp) => {
        let v = f(dp);
        if (v == null || v <= cur) {
            return cur;
        }
        return v;
    }, -Infinity)
}

function getSum(p,ts,f) {
    return ts.reduce((cur, dp) => {
        let v = f(dp);
        if (v == null) return cur;
        return cur + v;
    }, 0)
}

function getVar(p,ts,f) {
    return ts.reduce((cur, dp) => {
        let v = f(dp);
        if (v == null) return cur;
        return cur + v * v;
    }, 0);
}

function getNonNull(p,ts,f) {
    return ts.reduce((cur, dp) => (f(dp) == null ? cur : cur + 1), 0);
}

function getMean(d,ts,f) {
    return d.sum() / d.nonNull();
}

function getStdev(d,ts,f) {
    const mu = d.mean();
    return Math.sqrt(ts.reduce((cur, dp) => {
      let v = f(dp);
      if (v == null) return cur;
      return cur + Math.pow(v - mu, 2)
    }, 0) / (d.nonNull() - 1))
}

function registerAnalysis(ts) {
    ts.addAnalysisFunction("type",getType);
    ts.addAnalysisFunction("keys",getKeys);
    ts.addAnalysisFunction("min",getMin);
    ts.addAnalysisFunction("max",getMax);
    ts.addAnalysisFunction("sum",getSum);
    ts.addAnalysisFunction("var",getVar);
    ts.addAnalysisFunction("nonNull",getNonNull);
    ts.addAnalysisFunction("mean",getMean);
    ts.addAnalysisFunction("stdev",getStdev);
}

export default registerAnalysis;