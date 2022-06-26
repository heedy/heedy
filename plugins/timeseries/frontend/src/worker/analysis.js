import {
    createLTTB
} from "../../dist/downsample.mjs";

import {addMilliseconds} from "date-fns";

function cleanDT(ts) {
    // TODO: When merging data from multiple timeseries, heedy might not enforce
    // that there is no timestamp overlap. This function fixes the issue.
    for (let i = 0; i < ts.length - 1; i++) {
        if (ts[i].dt !== undefined && ts[i].t + ts[i].dt > ts[i + 1].t) {
            ts[i].dt = ts[i + 1].t - ts[i].t;
        }
    }
    return ts;
}


function getType(p, ts, extractor) {
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

function getKeys(p, ts, f) {
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


function getMin(p, ts, f) {
    return ts.reduce((cur, dp) => {
        let v = f(dp);
        if (v == null || v >= cur) {
            return cur;
        }
        return v;
    }, Infinity)
}

function getMax(p, ts, f) {
    return ts.reduce((cur, dp) => {
        let v = f(dp);
        if (v == null || v <= cur) {
            return cur;
        }
        return v;
    }, -Infinity)
}

function getSum(p, ts, f) {
    return ts.reduce((cur, dp) => {
        let v = f(dp);
        if (v == null) return cur;
        return cur + v;
    }, 0)
}

function getVar(p, ts, f) {
    return ts.reduce((cur, dp) => {
        let v = f(dp);
        if (v == null) return cur;
        return cur + v * v;
    }, 0);
}

function getNonNull(p, ts, f) {
    return ts.reduce((cur, dp) => (f(dp) == null ? cur : cur + 1), 0);
}

function getMean(d, ts, f) {
    return d.sum() / d.nonNull();
}

function getStdev(d, ts, f) {
    const mu = d.mean();
    return Math.sqrt(ts.reduce((cur, dp) => {
        let v = f(dp);
        if (v == null) return cur;
        return cur + Math.pow(v - mu, 2)
    }, 0) / (d.nonNull() - 1))
}

function getDate(d, ts, f) {
    return ts.map(dp => {
        const dv = f(dp);
        if (dv !== null) return new Date(dv * 1000);
        return null;
    });
}

function dateTransform(arr) {
    return arr.map(d => dv != null ? new Date(dv) : null);
}

function extract(arr, accesor) {
    return arr.map(dp => {
        const out = accesor(dp);
        return out !== undefined ? out : null;
    });
}

// filterNullTransform 
function filterNullTransform(arr, accessor) {
    const out = arr.filter(dp => accessor(dp) != null);
    // If it doesn't actually filter anything, return the original array
    // to save some memory when caching.
    if (out.length==arr.length) return arr;
    return out;
}

/**
 * Splits datapoints with durations into two elements - one at start of the duration,
 * and one at end of the duration
 *
 * @param {*} ts timeseries
 */
function explicitDuration(ts, options = {}) {
    options = {offset: 0.000,all:true, ...options};
    const res = new Array(ts.length * (options.separator !==undefined ? 3 : 2));
    let j = 0;
    for (let i = 0; i < ts.length; i++) {
        res[j] = ts[i];
        j++;
        if (options.all || (ts[i].dt !== undefined && ts[i].dt != 0)) {
            let v = ts[i].dt - options.offset;
            if (v===undefined || isNaN(v) || v < 0) v=0;
            res[j] = {
                t: addMilliseconds(ts[i].t,1000*v),
                d: ts[i].d,
                m: ts[i].m
            };
            j++;
        }
        if (options.separator!==undefined) {
            res[j] = options.separator;
            j++;
        }
    }
    return res.slice(0, j);
}

function downsample(ts, accessor, samples) {
    return createLTTB({
        x: (dp) => dp.t.getTime(),
        y: accessor
    })(ts, samples);
}


function registerAnalysis(ts) {
    ts.addAnalysisFunction("type", getType);
    ts.addAnalysisFunction("keys", getKeys);
    ts.addAnalysisFunction("min", getMin);
    ts.addAnalysisFunction("max", getMax);
    ts.addAnalysisFunction("sum", getSum);
    ts.addAnalysisFunction("var", getVar);
    ts.addAnalysisFunction("nonNull", getNonNull);
    ts.addAnalysisFunction("mean", getMean);
    ts.addAnalysisFunction("stdev", getStdev);
    ts.addAnalysisFunction("toDates", getDate);

    ts.addTransformFunction("toDates", dateTransform);
    ts.addTransformFunction("explicitDuration", explicitDuration);
    ts.addTransformFunction("filterNull", filterNullTransform, {
        accessor: true
    });
    ts.addTransformFunction("extract", extract, {
        accessor: true
    });
    ts.addTransformFunction("downsample", downsample, {
        accessor: true,
        extra_args: 1
    });
}

export default registerAnalysis;