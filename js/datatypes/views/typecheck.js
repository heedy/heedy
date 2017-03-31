/**
 * This file defines functions which check for various data formats for data arrays. The views use these functions to determine whether or not
 * they can do a visualization of the data.
 * 
 * These functions also cache their last argument, meaning that calls from multiple views to determine properties of a dataset are *very* cheap.
 */

// Check if the supplied data is of the given types
export const isBool = (d) => (typeof (d) === "boolean" || d === "true" || d === "false" || d === 1 || d === 0);
export const isNumber = (d) => (!isNaN(parseFloat(d)) && isFinite(d) || isBool(d));
export const isString = (d) => (typeof (d) === "string");
export const isObject = (d) => (typeof (d) === "object" && d !== null);
export const isKey = (d) => (isNumber(d) || isString(d));
export const isLocation = (d) => (isObject(d) && isNumber(d['latitude']) && isNumber(d['longitude']) && (d['accuracy'] === undefined || isNumber(d['accuracy'])));


// Returns true or false
function getBool(d) {
    if (typeof (d) === "boolean") {
        return d;
    }
    return (d === "true" || d === 1);
}

export function getNumber(d) {
    let n = parseFloat(d);
    if (isNaN(n)) return (getBool(d) ? 1 : 0);
    return n;
}

// I represents the data portion of a datapoint (identity). It is the default transform function used.
export const I = (d) => d.d;


// https://stackoverflow.com/questions/9716468/is-there-any-function-like-isnumeric-in-javascript-to-validate-numbers
function isOnlyNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

// This is a custom comparison function used to sort the keys in increasing order.
// We order things as follows:
//  - If we think that both keys are in a similar format, and have floats in them, sort by the float.
//  - Otherwise, perform a normal compare
var floatmatcher = /[+-]?\d+(\.\d+)?/g;
export function dataKeyCompare(a, b) {
    // We first try to extract a number from both strings
    // http://stackoverflow.com/questions/17374893/how-to-extract-floating-numbers-from-strings-in-javascript
    let numa = a.match(floatmatcher)
    if (numa != null && numa.length > 0) {
        let numb = b.match(floatmatcher)
        if (numb != null && numa.length == numb.length) {
            let na = parseFloat(numa[0]);
            let nb = parseFloat(numb[0]);
            return (na < nb
                ? -1
                : (na == nb
                    ? 0
                    : 1));
        }
    }

    // Since we couldn't extract a number, try to match the data values
    if (isOnlyNumeric(this[a]) && isOnlyNumeric(this[b])) {
        a = this[a];
        b = this[b];
    }

    // Otherwise, return just normal string compare
    return (a > b
        ? -1
        : (a == b
            ? 0
            : 1));
}


/**
 * The cache allows us to do expensive type checking. Each check is only done ONCE for
 * each datapoint array, with given transform function.
 * 
 * The cache has as its keys the datapoint arrays.
 * The value stored is another WeakMap of the transform functions. A transform function in this context
 * is NOT a pipescript transform, but rather a javascript function that transforms a datapoint into another
 * value. The specific details are left to the type-checking functions to figure out
 */
let typecache = new WeakMap();

/**
 * Gets the cached results from previous type checks
 * @param {array} d the datapoint array
 * @param {function} f the transform function(s) for this cache
 */
function getCache(d, f) {
    if (!typecache.has(d)) return {};
    return typecache.get(d).get(f) || {};
}

/**
 * Sets the cache 
 * @param {*} d the datapoint array
 * @param {*} f the transform function(s)
 * @param {*} v the object of values to set
 */
function setCache(d, f, v) {
    let c = {
        ...getCache(d, f),
        ...v
    };

    if (!typecache.has(d)) {
        let fcache = new WeakMap();
        fcache.set(f, c);
        typecache.set(d, fcache);
        return;
    }
    typecache.get(d).set(f, c);
}



/**
 * This function allows us to quickly return results without recomputing
 * things, which can be a pretty expensive operation once the datapoint arrays 
 * get large.
 * 
 * @param {string} key the key to use in the cache for storing results
 * @param {function} fn the function that computes the data
 */
function cacheWrapper(key, fn) {
    return function (d, f = I, nolog = false) {
        // if we already computed on this object, just return it.
        let c = getCache(d, f);
        if (c[key] !== undefined) {
            return c[key];
        }

        let ret = fn(d, f);
        let v = {};
        v[key] = ret;

        setCache(d, f, v);
        if (!nolog) console.log("TYPECHECK", key, ret);

        return ret;
    }
}




/**
 * Checks if the given datapoint array is all objects, and returns the keys of the object
 * as well as the min and max keys
 * @param {*} d datapoint array
 * @param {*} f transform function to use when checking
 */
export const keys = cacheWrapper('keys', function (d, f) {
    if (d.length === 0) return null;

    // This map will count the number of times each key is seen in the full array
    let keymap = {};

    for (let i = 0; i < d.length; i++) {
        let dp = f(d[i]);
        if (!isObject(dp)) return null;

        Object.keys(dp).map(function (k) {
            if (keymap[k] === undefined) {
                keymap[k] = 1;
            } else {
                keymap[k]++;
            }
        });
    }

    let mink = {};

    Object.keys(keymap).map(function (k) {
        if (keymap[k] === d.length) {
            mink[k] = true;
        }
    });

    // Return both the full keymap, and the min keymap
    // representing keys shared by all
    return {
        all: keymap,
        min: mink
    };
});


/**
 * Checks if the given datapoint sequence can be interpreted as a sequence of numbers,
 * and returns the transform function that allows one to extract the numbers.
 * 
 * @param {*} d datapoint array
 * @param {*} f transform function to use when checking. Default is identity.
 * 
 * @return transform function that gives the number, or null if not numeric
 */
export const numeric = cacheWrapper('numeric', function (d, f) {
    if (d.length === 0) return null;

    let key = "";
    // We want to be able to recognize an object with a single element which is a number as numeric,
    // since this is a very common case for T-Datasets.
    if (isObject(f(d[0]))) {
        // Looks like it is an object. If it has only one key, we can still use that one key
        let k = keys(d, f);
        if (k == null || Object.keys(k.all).length != 1 || Object.keys(k.min).length != 1) return null;
        key = Object.keys(k.min)[0];
        if (isObject(f(d[0])[key])) return null;
        let g = f;
        f = (x) => g(x)[key];
    }

    // Gets the info on the data elements
    let allbool = true;
    let allint = true;
    let min = 9999999999;
    let max = -min;
    for (let i = 0; i < d.length; i++) {
        let dp = f(d[i]);
        if (!isNumber(dp)) return null;
        if (allbool && !isBool(dp)) {
            allbool = false;
        }
        let n = getNumber(dp);
        if (allint && !Number.isInteger(n)) {
            allint = false;
        }
        if (n > max) max = n;
        if (n < min) min = n;
    }

    return {
        key: key,
        allbool: allbool,
        allint: allint,
        min: min,
        max: max,
        normalizer: min == max ? (d) => 0 : ((d) => (d - min) / (max - min)),
        f: (n) => getNumber(f(n))
    };
});

/**
 * Returns whether the data given can be considered categorical
 */
export const categorical = cacheWrapper('categorical', function (d, f) {
    if (d.length === 0) return null;

    let kv = new Map();
    let unique = 0;
    for (let i = 0; i < d.length; i++) {
        let dp = f(d[i]);
        if (!isKey(dp)) return null;
        if (!kv.has(dp)) {
            kv.set(dp, 0);
            unique++;
            if (unique > 200) return null;
        }
        kv.set(dp, kv.get(dp) + 1);
    }
    let v = {
        categories: unique,
        total: d.length,
        categorymap: kv
    };
    let c = (v.categories / v.total < 0.5 || v.categories < v.total && v.categories < 20);
    return c ? v : null;
});

export const location = cacheWrapper('location', function (d, f) {
    if (d.length === 0) return null;

    for (let i = 0; i < d.length; i++) {
        let dp = f(d[i]);
        if (!isLocation(dp)) return null;
    }
    return {
        boundingBox: true // TODO: a bounding box
    };
});

export const objectvalues = cacheWrapper('object', function (d, f) {
    if (d.length === 0) return null;
    if (location(d, f) !== null) return null;  // If it is a location, we pretend it is not an object anyore

    let k = keys(d, f);
    if (k === null) return null;

    let v = {};
    Object.keys(k.min).map(function (k) {
        let f2 = (x) => f(x)[k];
        v[k] = {
            numeric: numeric(d, f2, true),
            categorical: categorical(d, f2, true),
            location: location(d, f2, true)
        };
    });
    return v;
});