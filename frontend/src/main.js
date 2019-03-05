import Vue from "vue";
import VueRouter from "vue-router";
import Vuex from "vuex";

import Theme from "./cdb/theme.mjs";
import NotFound from "./cdb/404.mjs";
import Loading from "./cdb/loading.mjs";

// Add the two libraries
Vue.use(VueRouter);
Vue.use(Vuex);

// Add the app's routes to the router, with pages loaded dynamically
export const router = new VueRouter({
  routes: Object.keys(appinfo.routes)
    .map(k => ({
      path: k,
      component: () => import("./" + appinfo.routes[k]),
      loading: Loading,
      error: NotFound,
      delay: 200
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
  state: appinfo
});
// Vue is used as a global
export const vue = new Vue({
  router: router,
  store: store,
  render: h => h(Theme)
});

// Mount it
vue.$mount("#app");
