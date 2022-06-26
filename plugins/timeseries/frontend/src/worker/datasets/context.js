import {
    deepEqual
} from "../../../util.mjs";
import {
    stableStringify
} from "../../../dist/json-json-template.mjs";
const analysisFunctions = new Map();

const analysisFunction_defaultOptions = {
    cache: true,
    isTransform: false
};

function addAnalysisFunction(name, f, options = {}) {
    options = {
        ...analysisFunction_defaultOptions,
        ...options
    }
    analysisFunctions.set(name, {
        f,
        options
    });
}

const datasetContextInitializers = [];

function initDatasetContext(name, f) {
    datasetContextInitializers.push(f);
}

const getkey = (element, path, call, args) => `${element}[${path.join('.')}].${call}(${args.join(',')})`

function getAccessorFunction(path) {
    if (path.length == 0) return (d) => d;
    return (d) => {
        let v = d;
        for (let i = 0; i < path.length; i++) {
            v = v ?.[path[i]];
        }
        return v === undefined ? null : v;
    };
}

const transformFunctions = {};
const transformCache = new WeakMap();

function addTransformFunction(name, f, options = {}) {
    transformFunctions[name] = (arr, ...args) => {
        const key = `${name}${stableStringify(args)}`;
        let cachedValues = null;
        if (transformCache.has(arr)) {
            cachedValues = transformCache.get(arr);
            if (cachedValues.has(key)) {
                console.vlog(`Transform (cached) - ${key}`);
                return cachedValues.get(key);
            }
        } else {
            cachedValues = new Map();
            transformCache.set(arr, cachedValues);
        }
        if (options.accessor !== undefined && options.accessor) {
            const extra_args = options.extra_args!==undefined?options.extra_args:0;
            args = [getAccessorFunction(args.slice(extra_args)), ...args.slice(0, extra_args)];
        }
        console.vlog(`Transform - ${key}`);
        const data = f(arr, ...args);
        cachedValues.set(key, data);
        return data;
    }
}

const tpls = (...strings) => strings.map(s => "'" + s.replace("'", "\\'") + "'").join(",");



function analysisAccessor(cache, arr, element, path,key,idx) {
    const accessor = (dp) => {
        let v = dp[element];
        for (let i = 0; i < path.length; i++) {
            v = v ?.[path[i]];
        }
        return v === undefined ? null : v;
    }
    const api = (...args) => {
        if (args.length == 0) {
            
            const key = getkey(element, path, '', []);
            if (cache.has(key)) {
                console.vlog(`GET (cached) - ${element}.${path.join('.')}()`);
                return cache.get(key);
            }
            console.vlog(`GET ${element}.${path.join('.')}()`);
            const data = arr.map(accessor);
            cache.set(key, data);
            return data;
        }
        // Otherwise, it is a prop accessor, so we add them
        return analysisAccessor(cache, arr, element, path.concat(args),key,idx);
    }
    let pxy = null;
    pxy = new Proxy(api, {
        get(target, prop) {
            if (prop==="tplAccessor") {
                return (useKey=false) => {
                    let dsi = `[${idx}]`;
                    if (useKey) dsi = `.get(${tpls(key)})`;
                    if (path.length==0) return `data${dsi}.${element}`;
                    return `data[${dsi}].${element}(${tpls(...path)})`;
                }
            }
            if (prop==="path") {
                return path;
            }
            if (prop==="element") {
                return element;
            }
            if (prop==="key") {
                return key;
            }
            if (prop==="index") {
                return idx;
            }
            if (prop==="length") {
                return arr.length;
            }
            const res = analysisFunctions.get(prop);
            if (res !== undefined) {
                const {
                    f,
                    options
                } = res;
                if (options.cache) {
                    return (...args) => {
                        const key = getkey(element, path, prop, args);
                        if (cache.has(key)) {
                            console.vlog(`Analysis (cached) ${element}.${path.join('.')}.${prop}`);
                            return cache.get(key);
                        }
                        console.vlog(`Analysis ${element}.${path.join('.')}.${prop}`);
                        const data = f(pxy, arr, accessor, ...args);
                        cache.set(key, data);
                        return data;
                    }
                }
                console.vlog(`Analysis ${element}.${path.join('.')}.${prop}`);
                return (...args) => f(pxy, arr, accessor, ...args);
            }
            // This allows to get elements by index.
            if (arr[prop] === undefined) {
                throw new Error(`${element}.${path.join('.')} does not have property ${prop}`);
            }
            return accessor(arr[prop]);
        }
    });
    return pxy;
}

function addAnalysisAPI(arr,key,idx) {
    const cache = new Map();

    arr.d = analysisAccessor(cache, arr, 'd', [],key,idx);
    arr.m = analysisAccessor(cache, arr, 'm', [],key,idx);
    arr.t = analysisAccessor(cache, arr, 't', [],key,idx);
    arr.dt = analysisAccessor(cache, arr, 'dt', [],key,idx);

    arr.key = key;
    arr.index = idx;

    return arr;
}

function datasetAPI(ctx, d) {
    const data = new Map(Object.entries(d));

    data.values = ctx.keys.map((k) => d[k]);
    data.map = (f, ta) => data.values.map(f, ta);
    data.filter = (f, ta) => data.values.filter(f, ta);
    data.every = (f, ta) => data.values.every(f, ta);
    data.some = (f, ta) => data.values.some(f, ta);
    data.length = data.values.length;
    data.minTimestamp = () => {
        let min = (new Date()).getTime();
        data.values.forEach((arr) => {
            if (arr.length > 0) {
                min = Math.min(min, arr[0].t);
            }
        });
        return new Date(min);
    }
    data.maxTimestamp = () => {
        let max = 0;
        data.values.forEach((arr) => {
            if (arr.length > 0) {
                max = Math.max(max, arr[arr.length - 1].t);
            }
        });
        return new Date(max);
    }


    // Order the data elements alphabetically by keys
    ctx.keys.forEach((k, i) => {
        const arr = d[k];
        data[i] = arr;
        addAnalysisAPI(arr,k,i);
    });

    return data;
}

// Get the timeseries IDs that are included in an element of the query object
function getQueryElementTimeseries(elem) {
    let qobj = {};
    if (elem.timeseries !== undefined) {
        qobj[elem.timeseries] = true;
    }
    if (elem.merge !== undefined) {
        elem.merge.forEach((e2) => {
            qobj = {
                ...qobj,
                ...getQueryElementTimeseries(e2),
            };
        });
    }
    if (elem.dataset !== undefined) {
        Object.values(elem.dataset).forEach((e2) => {
            qobj = {
                ...qobj,
                ...getQueryElementTimeseries(e2),
            };
        });
    }
    return qobj;
}

function queryElementAPI(ctx, key, qe) {
    const qElement = {
        ...qe
    };
    qElement.isSimple = () => {
        if (qe.transform !== undefined && qe.transform.length > 0) {
            return false;
        }
        if (qe.post_transform !== undefined && qe.post_transform.length > 0) {
            return false;
        }
        if (qe.timeseries === undefined) {
            return false;
        }
        if (qe.dataset !== undefined) {
            return false;
        }
        return true;
    };
    qElement.getAllTimeseries = () => getQueryElementTimeseries(qe);
    return qElement;
}

function CleanQuery(q) {
    let q2 = {};
    Object.keys(q).forEach((k) => {
        let e = q[k];
        let e2 = {
            ...q[k],
        };

        if (e.i1 !== undefined && !isNaN(e.i1)) {
            e2.i1 = parseInt(e.i1);
        }
        if (e.i2 !== undefined && !isNaN(e.i2)) {
            e2.i2 = parseInt(e.i2);
        }
        if (e.limit !== undefined && !isNaN(e.limit)) {
            e2.limit = parseInt(e.limit);
        }
        if (e.i !== undefined && !isNaN(e.i)) {
            e2.i = parseInt(e.i);
        }
        q2[k] = e2;
    });
    return q2;
}

function queryAPI(ctx, q) {
    q = CleanQuery(q);
    const query = new Map();
    ctx.keys.map((k, i) => {
        const qElement = queryElementAPI(ctx, k, q[k]);
        query[i] = qElement;
        query.set(k, qElement);
    });

    query.isEqual = (q2) => deepEqual(q, CleanQuery(q2));
    query.isAlphabeticallyEqual = (q2) => {
        q2 = CleanQuery(q2);
        const q1a = ctx.keys.map(k => q[k]);
        const q2a = Object.keys(q2).map(k => q2[k]);
        return deepEqual(q1a, q2a);
    }

    return query;
}




class DatasetContext {
    constructor(query, dataset, timeseries, settings) {
        // A sorted list of keys in the dataset
        this.keys = Object.keys(dataset).sort();

        this.timeseries = timeseries;
        this.settings = settings;

        this.query = queryAPI(this, query);
        this.data = datasetAPI(this, dataset);

        // Add transform functions - functions that have an array as first
        // argument, and return a transform of the array.
        Object.keys(transformFunctions).forEach((k) => {
            this[k] = transformFunctions[k]
        });


        // Allow custom things added to the context object
        datasetContextInitializers.forEach((f) => f(this));
    }

    // A simple simplifying code for generating template elements
    tpl(expr) {
        return "${{" + expr + "}}"
    }

    tpls(...strings) {
        return tpls(...strings);
    }

    getSeriesLabelTemplate(idx,key="") {
        if (key.length>0) {
            key = ` (${key})`
        }
        if (typeof idx==="string") {
            if (!this.query.get(idx).isSimple()) {
                return idx +key;
            }
            return this.tpl(`timeseries[query.get(${this.tpls(idx)}).timeseries].name`) +key;
        }
        if (!this.query[idx].isSimple()) {
            return this.keys[idx] + key; // On complex queries, just use the key
        }
        return this.tpl(`timeseries[query[${idx}].timeseries].name`) +key
    }

}

export {
    addAnalysisFunction,
    addTransformFunction,
    initDatasetContext,
    getQueryElementTimeseries
};
export default DatasetContext;