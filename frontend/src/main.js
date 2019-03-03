import Vue from "vue";
import VueRouter from "vue-router";

import Theme from "./js/theme.jsm";

// Add the vue router
Vue.use(VueRouter);

// Add the app's routes to the router, with pages loaded dynamically
const router = new VueRouter({
  routes: Object.keys(appinfo.routes).map(k => ({
    path: k,
    component: () => import("./" + appinfo.routes[k])
  }))
});

// Vue is used as a global
new Vue({
  router: router,
  render: h => h(Theme)
}).$mount("#app");
