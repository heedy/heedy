import {deepEqual} from "../../../util.mjs";

const analysisFunctions = new Map();

function addAnalysisFunction(name, f, cacheResult = true) {
    analysisFunctions.set(name, {f,cacheResult});
}

const datasetContextInitializers = [];

function initDatasetContext(name, f) {
    datasetContextInitializers.push(f);
}

const getkey = (element, path, call, args) => `${element}[${path.join('.')}].${call}(${args.join(',')})`

function analysisAccessor(cache, arr, element, path) {
    const accessor = (dp) => {
        let v = dp[element];
        for (let i = 0; i < path.length; i++) {
            v = v?.[path[i]];
        }
        return v === undefined ? null : v;
    }
    const api = (...args) => {
        if (args.length == 0) {
            console.log(`GET ${element}.${path.join('.')}()`);
            const key = getkey(element, path, '', []);
            if (cache.has(key)) {
                console.log(`USING CACHED ${element}.${path.join('.')}()`);
                return cache.get(key);
            }
            const data = arr.map(accessor);
            cache.set(key, data);
            return data;
        }
        // Otherwise, it is a prop accessor, so we add them
        return analysisAccessor(cache, arr, element, path.concat(args));
    }

    return new Proxy(api, {
        get: (target, prop) => {
            const {
                f,
                cacheResult
            } = analysisFunctions.get(prop);
            if (f !== undefined) {
                if (cacheResult) {
                    return (...args) => {
                        const key = getkey(element, path, prop, args);
                        if (cache.has(key)) {
                            console.log(`Analysis (cached) ${element}.${path.join('.')}.${prop}`);
                            return cache.get(key);
                        }
                        console.log(`Analysis ${element}.${path.join('.')}.${prop}`);
                        const data = f(arr, accessor, ...args);
                        cache.set(key, data);
                        return data;
                    }
                }
                console.log(`Analysis ${element}.${path.join('.')}.${prop}`);
                return (...args) => f(arr, accessor, ...args);
            }
            // This allows to get elements by index.
            return accessor(arr[prop]);
        }
    });
}

function addAnalysisAPI(arr) {
    const cache = new Map();

    arr.d = analysisAccessor(cache, arr, 'd', []);
    arr.m = analysisAccessor(cache, arr, 'm', []);
    arr.t = analysisAccessor(cache, arr, 't', []);
    arr.dt = analysisAccessor(cache, arr, 'dt', []);

    return arr;
}

function datasetAPI(ctx,d) {
    const data = new Map(Object.entries(d));

    data.values = ctx.keys.map((k)=> d[k]);

    // Order the data elements alphabetically by keys
    ctx.keys.forEach((k,i)=> {
        const arr = d[k];
        data[i] = arr;
        addAnalysisAPI(arr);
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
          ...getQueryElementObjects(e2),
        };
      });
    }
    if (elem.dataset !== undefined) {
      Object.values(elem.dataset).forEach((e2) => {
        qobj = {
          ...qobj,
          ...getQueryElementObjects(e2),
        };
      });
    }
    return qobj;
  }

function queryElementAPI(ctx,key,qe) {
    const qElement = {...qe};
    qElement.isSimple = () => {
        if (qe.transform!==undefined && qe.transform.length>0) {
            return false;
        }
        if (qe.post_transform!==undefined && qe.post_transform.length>0) {
            return false;
        }
        if (qe.timeseries===undefined) {
            return false;
        }
        if (qe.dataset!==undefined) {
            return false;
        }
        return true;
    };
    qElement.getAllTimeseries = () => {
        getQueryElementTimeseries(qe)
    }
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
  
function queryAPI(ctx,q) {
    q = CleanQuery(q);
    const query = new Map();
    ctx.keys.map((k,i) => {
        const qElement = queryElementAPI(ctx,k,q[k]);
        query[i] = qElement;
        query.set(k,qElement);
    });

    query.isEqual = (q2) => deepEqual(q,CleanQuery(q2));
    query.isAlphabeticallyEqual = (q2) => {
        q2 = CleanQuery(q2);
        const q1a = ctx.keys.map(k=> q[k]);
        const q2a = Object.keys(q2).map(k=> q2[k]);
        return deepEqual(q1a,q2a);
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
        this.data = datasetAPI(this,dataset);

        // Allow custom things added to the context object
        datasetContextInitializers.forEach((f) => f(this));
    }

    // A simple simplifying code for generating template elements
    tpl(expr) {
        return "${{" + expr + "}}"
    }

    tpls(...strings) {
        strings.map(s=> "'" + s.replace("'","\\'") + "'").join(",")
    }

}

export {
    addAnalysisFunction,
    initDatasetContext,
    getQueryElementTimeseries
};
export default DatasetContext;