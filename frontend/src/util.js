// https://stackoverflow.com/questions/1714786/query-string-encoding-of-a-javascript-object
export function urlify(obj) {
  var str = [];
  for (var p in obj)
    if (obj.hasOwnProperty(p)) {
      str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
    }
  return str.join("&");
}

/**
 * This function allows querying the API explicitly. If the method is get, data is urlencoded.
 * It explicitly returns the resulting object, or throws the error given
 * @param {string} method - HTTP verb to use (GET/POST/...)
 * @param {string} uri - uri to query (api/heedy/...)
 * @param {object} data - optional object to send as a json payload (or FormData), or as url params if GET/DELETE request
 * @param {object} opt - options to use for the request
 */
async function api(method,uri,body=null,opt={}) {
  const fetch_options = Object.assign({
    method: method,
    credentials: "include",
    redirect: "follow",
    headers: Object.assign({},opt.headers)
  },opt.fetch);
  if (opt.signal!==undefined) {
    fetch_options.signal = opt.signal;
  }

  const setContentType = (content_type) => {
    if (fetch_options.headers['Content-Type'] === undefined) {
      fetch_options.headers['Content-Type'] = content_type;
    }
  };

  const urlparams = Object.assign({},opt.params);

  // Figure out how to encode the body if it is data
  if (body!=null) {
    if (body instanceof FormData) {
      fetch_options.body = body;
    } else {
      if (opt.type === "form-data") {
        const fd = new FormData();
        for (let key in body) {
          fd.append(key, body[key]);
        }
        fetch_options.body = fd;
      } else if (opt.type === "urlencoded") {
        fetch_options.body = urlify(body);
        setContentType("application/x-www-form-urlencoded");
      } else if (method=="GET" || method=="DELETE") {
        // For simplified use of the API, GET requests send body as url params
        Object.assign(urlparams,body);
      } else {
        fetch_options.body = JSON.stringify(body);
        setContentType("application/json");
      }
    }
  }

  if (Object.keys(urlparams).length>0) {
    uri = uri + (uri.indexOf("?")>=0?'&':'?')  + urlify(urlparams);
  }

  try {
    var resp = await fetch(uri, fetch_options);
  } catch (err) {
    return {
      response: {
        ok: false
      },
      data: {
        error: "fetch_error",
        error_description: "Could not connect to the server",
        id: "?"
      }
    };
  }
  if (opt.output_type==="raw") {
    // If the response isn't 
    return {response: resp,data: resp.body};
  }

  try {
    return {
      response: resp,
      data: await resp.json()
    };
  } catch (err) {
    return {
      response: resp,
      data: {
        error: "response_error",
        error_description: err.message,
        id: "?"
      }
    };
  }

}



//https://stackoverflow.com/questions/30476150/javascript-deep-comparison-recursively-objects-and-properties
function getClass(obj) {
  return Object.prototype.toString.call(obj);
}

/*
 ** @param a, b        - values (Object, RegExp, Date, etc.)
 ** @returns {boolean} - true if a and b are the object or same primitive value or
 **                      have the same properties with the same values
 */
function deepEqual(a, b) {
  // If a and b reference the same value, return true
  if (a === b) return true;

  // If a and b aren't the same type, return false
  if (typeof a != typeof b) return false;

  // Already know types are the same, so if type is number
  // and both NaN, return true
  if (typeof a == "number" && isNaN(a) && isNaN(b)) return true;

  // Get internal [[Class]]
  var aClass = getClass(a);
  var bClass = getClass(b);

  // Return false if not same class
  if (aClass != bClass) return false;

  // If they're Boolean, String or Number objects, check values
  if (
    aClass == "[object Boolean]" ||
    aClass == "[object String]" ||
    aClass == "[object Number]"
  ) {
    if (a.valueOf() != b.valueOf()) return false;
  }

  // If they're RegExps, Dates or Error objects, check stringified values
  if (
    aClass == "[object RegExp]" ||
    aClass == "[object Date]" ||
    aClass == "[object Error]"
  ) {
    if (a.toString() != b.toString()) return false;
  }

  // For functions, check stringigied values are the same
  // Almost impossible to be equal if a and b aren't trivial
  // and are different functions
  if (aClass == "[object Function]" && a.toString() != b.toString())
    return false;

  // For all objects, (including Objects, Functions, Arrays and host objects),
  // check the properties
  var aKeys = Object.keys(a);
  var bKeys = Object.keys(b);

  // If they don't have the same number of keys, return false
  if (aKeys.length != bKeys.length) return false;

  // Check they have the same keys
  if (
    !aKeys.every(function (key) {
      return b.hasOwnProperty(key);
    })
  )
    return false;

  // Check key values - uses ES5 Object.keys
  return aKeys.every(function (key) {
    return deepEqual(a[key], b[key]);
  });
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export { deepEqual, sleep };

export default async function consoleAPI(
  method,
  uri,
  data = null,
  options = {}
) {
  let res = await api(method, uri, data, options);
  if (!res.response.ok) {
    console.error(method, uri, data, options,res);
  }
  return res;
}