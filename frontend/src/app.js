import Vue, {
  VueRouter,
  Vuex,
  Vuetify
} from "./dist/vue.mjs";

import api from "./api.js";

import App from "./app/app.js";
import vuexStore from "./app/vuex.js";

async function setup(appinfo) {
  console.log("Setting up...");

  // Start running the import statements
  let plugins = appinfo.frontend.map(f => import("./" + f.path));

  // Prepare the vuex store
  const store = new Vuex.Store(vuexStore(appinfo));

  let app = new App(Vue, appinfo, store);

  app.api = api;

  for (let i = 0; i < plugins.length; i++) {
    console.log("Preparing", appinfo.frontend[i].name);
    try {
      (await plugins[i]).default(app);
    } catch (err) {
      console.error(err);
      alert(`Failed to load plugin '${appinfo.frontend[i].name}': ${err.message}`);
    }

  }

  // Now go through the injected modules to run their onInit
  for (let key in app.injected) {
    // skip loop if the property is from prototype
    if (!app.injected.hasOwnProperty(key)) continue;
    if (app.injected[key].$onInit !== undefined) {
      app.injected[key].$onInit();
    }
  }

  let routes = Object.values(app.routes);
  if (app.notFound !== null) {
    routes.push({
      path: "*",
      component: app.notFound
    })
  }

  // Set up the app routes
  const router = new VueRouter({
    routes: routes,
    // https://router.vuejs.org/guide/advanced/scroll-behavior.html#scroll-behavior
    scrollBehavior(to, from, savedPosition) {
      if (savedPosition) {
        return savedPosition
      } else {
        return {
          x: 0,
          y: 0
        }
      }
    }
  });

  const vuetify = new Vuetify({
    icons: {
      iconfont: 'md',
    },
  });

  Vue.mixin({
    computed: {
      $app() {
        return app;
      }
    }
  });

  const vue = new Vue({
    router: router,
    store: store,
    vuetify: vuetify,
    render: h => h(app.theme)
  });

  // Mount it
  vue.$mount("#app");

  return app;
}

export default setup;