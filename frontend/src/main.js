import Vue from "vue";
import VueRouter from "vue-router";
import Vuex, { mapState } from "vuex";

import Theme from "./heedy/theme.mjs";
import NotFound from "./heedy/404.mjs";
import Loading from "./heedy/loading.mjs";

// Add the two libraries
Vue.use(VueRouter);
Vue.use(Vuex);

// The vuex mapState
export { mapState };

// Add the app's routes to the router, with pages loaded dynamically

// TODO: https://github.com/vuejs/vue-router/pull/2140/commits/8975db3659401ef5831ebf2eae5558f2bf3075e1
// Lazy loading and error pages are not functional in router. Will need to fix this before release
export const router = new VueRouter({
  routes: Object.keys(appinfo.routes)
    .map(k => ({
      path: k,
      component: () => ({
        component: import("./" + appinfo.routes[k]),
        loading: Loading,
        error: NotFound,
        delay: 200,
        timeout: 0
      })
    }))
    .concat([
      {
        path: "*",
        component: NotFound
      }
    ])
});

// store is a global variable, since it can be used by external modules to add their own state management
export const store = new Vuex.Store({
  state: {
    info: appinfo,
    alert: {
      value: false,
      text: "",
      type: ""
    }
  },
  mutations: {
    alert(state, v) {
      state.alert = {
        value: true,
        type: "",
        text: "",
        ...v
      };
    }
  },
  actions: {
    errnotify({ commit }, v) {
      // Notifies of an error
      if (v.hasOwnProperty("error")) {
        // Only notify if it is an actual error
        commit("alert", {
          type: "error",
          text: v.error_description
        });
      }
    }
  }
});
// Vue is used as a global
export const vue = new Vue({
  router: router,
  store: store,
  render: h => h(Theme)
});

// Mount it
vue.$mount("#app");

// https://stackoverflow.com/questions/1714786/query-string-encoding-of-a-javascript-object
function urlify(obj) {
  var str = [];
  for (var p in obj)
    if (obj.hasOwnProperty(p)) {
      str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
    }
  return str.join("&");
}

// This function allows querying the API explicitly.
// If the method is get, data is urlencoded.
// It explicitly returns the resulting object, or throws the error given
export async function api(method, uri, data = null, json = true) {
  let options = {
    method: method,
    credentials: "include",
    redirect: "follow",
    headers: {}
  };
  if (data != null) {
    if (method == "GET") {
      uri = uri + "?" + urlify(data);
    } else if (json) {
      options.body = JSON.stringify(data);
      options.headers["Content-Type"] = "application/json";
    } else {
      options.body = urlify(data);
      options.headers["Content-Type"] = "application/x-www-form-urlencoded";
    }
  }
  console.log(uri, options);
  let response = await fetch(uri, options);
  console.log(response);
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
