import Vue from "vue";
import VueRouter from "vue-router";
import Vuex from "vuex";

import Theme from "./heedy/theme.mjs";
import NotFound from "./heedy/404.mjs";
import Loading from "./heedy/loading.mjs";

// Add the two libraries
Vue.use(VueRouter);
Vue.use(Vuex);

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
