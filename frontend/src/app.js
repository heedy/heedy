import Vue, {
  VueRouter,
  Vuex,
  Vuetify
} from "./dist.mjs";

var vuexPlugins = [];
var vuexModules = {};

var appMenu = {};

var injected = {};



// routes need pre-processing
var routes = {};

var currentTheme = null;

class App {
  constructor(appinfo, pluginName) {
    this.info = appinfo;
    this.pluginName = pluginName;

    // Add all injected subclasses to the global app object
    for (let key in injected) {
      // skip loop if the property is from prototype
      if (!injected.hasOwnProperty(key)) continue;
      this[key] = new injected[key](pluginName);
    }

  }

  /**
   * Adds a vuex module to the main app store.
   * 
   * @param {*} module Vuex module to add
   */
  addVuexModule(module) {
    vuexModules[this.pluginName] = module;
  }
  /**
   * Adds a vuex plugin to the main store.
   * @param {*} p plugin
   */
  addVuexPlugin(p) {
    vuexPlugins.push(p)
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
    routes[r.path] = r;
  }

  /**
   * The theme for the app.
   * @param {*} t vue object for the main theme
   */
  setTheme(t) {
    currentTheme = t;
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
    appMenu[m.key] = m;
  }


  inject(name, p) {
    injected[name] = p;
    this[name] = new injected[name](this.pluginName);
  }

}

async function setup(appinfo) {
  console.log("Setting up...");

  // Start running the import statements
  let plugins = appinfo.frontend.map(f => import("./" + f.path));

  for (let i = 0; i < plugins.length; i++) {
    console.log("Preparing", appinfo.frontend[i].name);
    try {
      (await plugins[i]).default(new App(appinfo, appinfo.frontend[i].name));
    } catch (err) {
      console.error(err);
      alert(`Failed to load plugin '${appinfo.frontend[i].name}': ${err.message}`);
    }

  }

  // Now go through the injected modules to run their onInit
  for (let key in injected) {
    // skip loop if the property is from prototype
    if (!injected.hasOwnProperty(key)) continue;
    (injected[key]["$onInit"] || (() => (1)))();
  }

  // There is a single built in vuex module, which holds 
  // the app info, the main menu, the extra menu, 
  // and other core information.
  vuexModules["app"] = {
    state: {
      info: appinfo,
      // menu_items gives all the defined menu items
      menu_items: appMenu,
    },
    mutations: {
      updateLoggedInUser(state, v) {
        state.info.user = v;
      }
    }
  };

  // Prepare the vuex store
  const store = new Vuex.Store({
    modules: vuexModules,
    plugins: vuexPlugins
  });

  // Set up the app routes
  const router = new VueRouter({
    routes: Object.values(routes),
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

  const vue = new Vue({
    router: router,
    store: store,
    vuetify: vuetify,
    render: h => h(currentTheme)
  })

  // Mount it
  vue.$mount("#app");


}

export default setup;