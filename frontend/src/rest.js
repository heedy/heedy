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
 * @param {object} data - optional object to send as a json payload
 * @param {object} params - params to set as url params
 * @param {boolean} json - whether data should be sent as standard POST url encoded or as json
 */
async function api(method, uri, data = null, params = null, json = true) {
  let options = {
    method: method,
    credentials: "include",
    redirect: "follow",
    headers: {}
  };
  if (params != null) {
    if (data != null && method == "GET") {
      uri =
        uri +
        "?" +
        urlify({
          ...data,
          ...params
        });
    } else {
      uri = uri + "?" + urlify(params);
    }
  }
  if (data != null) {
    if (method == "GET") {
      if (data != null && params == null) {
        uri = uri + "?" + urlify(data);
      }
    } else if (json) {
      options.body = JSON.stringify(data);
      options.headers["Content-Type"] = "application/json";
    } else {
      options.body = urlify(data);
      options.headers["Content-Type"] = "application/x-www-form-urlencoded";
    }
  }
  try {
    var response = await fetch(uri, options);
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

  try {
    return {
      response: response,
      data: await response.json()
    };
  } catch (err) {
    return {
      response: response,
      data: {
        error: "response_error",
        error_description: err.message,
        id: "?"
      }
    };
  }
}

export default async function consoleAPI(
  method,
  uri,
  data = null,
  params = null,
  json = true
) {
  let res = await api(method, uri, data, params, json);
  if (!res.response.ok) {
    console.error(method, uri, data, params, json, res);
  }
  return res;
}

var cssLinks = {};
var jsScripts = {};

export function addCSS(linkurl, integrity = "", crossorigin = "anonymous") {
  if (linkurl in cssLinks) return;
  cssLinks[linkurl] = true;
  var link = document.createElement("link");
  link.setAttribute("rel", "stylesheet");
  link.setAttribute("type", "text/css");
  link.setAttribute("href", linkurl);
  if (integrity != "") {
    link.setAttribute("integrity", integrity);
  }
  link.setAttribute("crossorigin", crossorigin);
  document.getElementsByTagName("head")[0].appendChild(link);
}

export function addJS(srcurl, integrity = "", crossorigin = "anonymous") {
  if (srcurl in jsScripts) return;
  jsScripts[srcurl] = true;
  var link = document.createElement("script");
  link.setAttribute("type", "application/javascript");
  link.setAttribute("src", srcurl);
  if (integrity != "") {
    link.setAttribute("integrity", integrity);
  }
  link.setAttribute("crossorigin", crossorigin);
  document.getElementsByTagName("head")[0].appendChild(link);
}
