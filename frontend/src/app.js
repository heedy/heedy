import Vue, {
  VueRouter,
  Vuex,
  Vuetify,
  createLogger
} from "./dist/vue.mjs";

import api from "./api.js";

import worker from "./worker.mjs"
class WorkerInjector {
  constructor() {
    this.handlers = {};

    worker.postMessage = (key, msg) => {
      return this._onMessage({
        data: {
          key: key,
          msg: msg
        }
      });
    }

    this.worker = {
      postMessage: (msg) => worker._onMessage({
        data: msg
      })
    };
  }
  addHandler(key, f) {
    this.handlers[key] = f;
  }

  /**
   * Sends a message with the given key to the worker
   * @param {*} key 
   * @param {*} msg 
   */
  postMessage(key, msg) {
    this.worker.postMessage({
      key: key,
      msg: msg
    });
  }

  add(filename) {
    this.postMessage("import", filename);
  }


  async _onMessage(e) {
    let msg = e.data;
    console.log("App:", msg);
    if (this.handlers[msg.key] !== undefined) {
      let ctx = {
        key: msg.key
      }
      await this.handlers[msg.key](ctx, msg.msg);
    } else {
      console.error(`Unknown message key ${msg.key}`);
    }
  }
}

class App {
  constructor(appinfo, store) {
    // Allows registration of components
    this.vue = Vue;
    // Allows setting stuff in the store
    this.store = store;

    this.info = appinfo;

    this.theme = null;
    this.injected = {};
    this.routes = {};


    this.worker = new WorkerInjector();
  }

  /**
   * Routes sets up the app's routes, one by one. It allows
   * for overriding routes, however, it only allows overriding the
   * base route, it does not handle child routes. To set up different
   * routes for logged in users and the public, the plugin can check
   * if info.user is null.
   * 
   * @param {*} r a single route element.
   */
  addRoute(r) {
    this.routes[r.path] = r;
  }


  /**
   * Add an item to the main app menu. 
   * 
   * @param {*} m Menu item to add. It is given an object
   *        with items "key", which is a unique ID, text, the text to display,
   *        icon, the icon to show, and route, which is the route to navigate to.
   *        Optionally also has a "location" attribute which hints at where the user
   *        might want the menu (primary,secondary,spaced_primary). 
   *        Can also have "component" which is a vue component to display instead of icon.
   *        Be aware that the component must have a "state" prop, where it is told how to behave
   *        i.e. whether the menu is small, on bottom, etc.
   */
  addMenuItem(m) {
    this.store.commit('addMenuItem', m);
  }

  inject(name, p) {
    this.injected[name] = p;
    this[name] = p;
  }

}

async function setup(appinfo) {
  console.log("Setting up...");

  // Start running the import statements
  let plugins = appinfo.frontend.map(f => import("./" + f.path));

  // Prepare the vuex store
  const store = new Vuex.Store({
    modules: {
      app: {
        state: {
          info: appinfo,
          // menu_items gives all the defined menu items
          menu_items: {},
        },
        mutations: {
          updateLoggedInUser(state, v) {
            state.info.user = v;
          },
          addMenuItem(state, m) {
            state.menu_items[m.key] = m;
          }
        }
      }
    },
    plugins: [createLogger()]
  });

  let app = new App(appinfo, store);

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

  // Set up the app routes
  const router = new VueRouter({
    routes: Object.values(app.routes),
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