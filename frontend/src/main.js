import Vue, { VueRouter, Vuex, Vuetify } from "./dist/vue.mjs";

import Frontend from "./main/frontend.js";
import WorkerInjector from "./main/worker_injector.js";
import WebsocketInjector from "./main/websocket.js";
import vuexStore from "./main/vuex.js";

async function setup(appinfo) {
  console.log("Setting up...", appinfo);

  // Start running the import statements
  let plugins = appinfo.plugins.map(f => import("./" + f.path));

  // Prepare the vuex store
  const store = new Vuex.Store(vuexStore(appinfo));

  let frontend = new Frontend(Vue, appinfo, store);

  // The websocket and worker come by default
  frontend.inject("worker", new WorkerInjector(appinfo));
  frontend.inject("websocket", new WebsocketInjector(frontend));

  for (let i = 0; i < plugins.length; i++) {
    console.log("Preparing", appinfo.plugins[i].name);
    try {
      (await plugins[i]).default(frontend);
    } catch (err) {
      console.error(err);
      alert(
        `Failed to load plugin '${appinfo.plugins[i].name}': ${err.message}`
      );
    }
  }

  // Now go through the injected modules to run their onInit
  for (let key in frontend.injected) {
    // skip loop if the property is from prototype
    if (!frontend.injected.hasOwnProperty(key)) continue;
    if (frontend.injected[key].$onInit !== undefined) {
      frontend.injected[key].$onInit();
    }
  }

  let routes = Object.values(frontend.routes);
  if (frontend.notFound !== null) {
    routes.push({
      path: "*",
      component: frontend.notFound
    });
  }

  // Set up the app routes
  const router = new VueRouter({
    routes: routes,
    // https://router.vuejs.org/guide/advanced/scroll-behavior.html#scroll-behavior
    scrollBehavior(to, from, savedPosition) {
      if (savedPosition) {
        return savedPosition;
      } else {
        return {
          x: 0,
          y: 0
        };
      }
    }
  });
  // Set the router in the frontend
  frontend.router = router;

  const vuetify = new Vuetify({
    icons: {
      iconfont: "md"
    }
  });

  Vue.mixin({
    computed: {
      $frontend() {
        return frontend;
      }
    }
  });

  const vue = new Vue({
    router: router,
    store: store,
    vuetify: vuetify,
    render: h => h(frontend.theme)
  });

  // Mount it
  vue.$mount("#frontend");

  return frontend;
}

export default setup;
