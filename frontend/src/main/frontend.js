import api from "../util.mjs";
/**
 * @alias frontend
 */
class Frontend {
  /**
   * The frontend object is passed to the setup functions of each plugin,
   * and exposes the APIs necessary to augment heedy's UI.
   * @example
   * function setup(frontend) {
   *    frontend.addRoute({
   *        path: "/myplugin/myroute",
   *        component: MyComponent
   *    });
   * }
   */
  constructor(Vue, appinfo, store) {
    /**
     * The `Vue <https://vuejs.org/>`_ instance. This is mainly to be used to register components and plugins.
     *
     * @example
     * frontend.vue.use(MyVuePlugin);
     * frontend.vue.component("mycomponent",MyComponent);
     */
    this.vue = Vue;

    /**
     * The `vuex <https://vuex.vuejs.org/>`_ store used for the frontend. A plugin
     * can set up its own `vuex modules <https://vuex.vuejs.org/guide/modules.html>`_
     * to manage its state.
     * @example
     * frontend.store.registerModule("myplugin",{
     *  state: {count: 0},
     *  mutations: {
     *      increment(state) {state.count++}
     *  }
     * });
     */
    this.store = store;

    /**
     * The `vue router <https://router.vuejs.org/>`_ used for the frontend. This value is only
     * initialized after the initial setup, so it is *not available* when initializing plugins.
     */
    this.router = null;

    /**
     * This property contains information passed in from the server, and is used
     * to set up the session.
     * @example
     * frontend.info = {
     *  // Whether the current session is of an admin user
     *  admin: true,
     *  // The list of plugins with frontend code,
     *  // as well as paths to their modules
     *  plugins: [
     *      {name: "heedy", path: "heedy/main.mjs"},
     *      ...
     *      {name: "timeseries", path: "timeseries/main.mjs"},
     *      {name: "myplugin", path: "myplugin/main.mjs"}
     *  ],
     *  // The currently logged in user. If no user is logged in,
     *  // this field is null. Checking if frontend.info.user
     *  // is null is the recommended way to modify plugin features
     *  // for logged in users.
     *  user: {
     *      username: "myuser",
     *      name: "My User",
     *      icon: "base64:...",
     *      public_read: false,
     *      users_read: false
     *  }
     * }
     */
    this.info = appinfo;

    /**
     * The vue component to use as the main theme for the frontend. The theme
     * renders the main menu, and holds the router that shows individual pages.
     *
     * @example
     * frontend.theme = MyThemeComponent
     */
    this.theme = null;

    /**
     * The vue component to display when linked to a route that was not registered.
     * For example, the notFound component will be displayed when the path `/#/blahblah`
     * is used.
     * @example
     * frontend.notFound = MyNotFoundComponent
     */
    this.notFound = null;

    this.injected = {};
    this.routes = {};
  }

  /**
   * An async helper function that allows querying the REST API explicitly.
   * It explicitly returns the decoded json object, or returns the error response.
   * @example
   * let username = "myuser"
   * let res = await frontend.rest("GET",`api/users/{encodeURIComponent(username)}`,{
   *  icon: true // urlparam to include icon with request
   * })
   * if (!res.response.ok) {
   *  console.log("Failed:",res.data.error_description);
   * } else {
   *  console.log("User:",res.data);
   * }
   *
   * @param {string} method - HTTP verb to use (GET/POST/...)
   * @param {string} uri - uri to query (api/users/...)
   * @param {object} data - optional object to send as a json payload. If the method is GET, the data is sent as url params.
   * @param {object} params - params to set as url params
   * @param {string} json - format of POST data, default is json. Can also be form-data or x-www-form-urlencoded
   *
   * @returns {object} an object with two fields, response and data. response gives a fetch query response object, and data contains the response content decoded from json.
   */
  async rest(method, uri, data = null, params = null, json = "json") {
    return await api(method, uri, data, params, json);
  }

  /**
   * Routes sets up the app's browser bar routing (portion of the path after /#/)
   * 
   * To set up different
   * routes for logged in users and the public, the plugin can check
   * if info.user is null.
   * 
   * @example
   * frontend.addRoute({
   *  path: "myplugin/mypath", // This means /#/myplugin/mypath
   *  component: MyComponent
   * });
   *
   * @param {string} r.path The path at which to define the route
   * @param {vue.Component} r.component The vue component object to show as the page at that route.
   */
  addRoute(r) {
    this.routes[r.path] = r;
  }

  /**
   * Add an item to the main app menu.
   *
   * @param {object} m Menu item to add.
   * @param {string} m.key a unique ID for the menu item
   * @param {string} m.text the text to display
   * @param {string} m.icon The icon to show
   * @param {string} m.route the route to navigate to on clicking the menu item
   * @param {string} [m.location] optional, hints at where the user
   *        might want the menu (primary,secondary,spaced_primary).
   * @param {vue.Component} [m.component] an optional vue component to display instead of icon.
   *        Be aware that the component must have a "state" prop, where it is told how to behave
   *        i.e. whether the menu is small, on bottom, etc.
   * @example
   * frontend.addMenuItem({
   *    key: "myMenuItem",
   *    text: "My Menu Item",
   *    icon: "fas fa-handshake",
   *    route: "/myplugin/mypath",
   *    location: "primary"
   * });
   */
  addMenuItem(m) {
    this.store.commit("addMenuItem", m);
  }

  /**
   * A plugin can inject its own API into the frontend object, so that all plugins loaded
   * after it have access to it.
   * @example
   * class MyAPI {
   *    constructor() {}
   *    myfunction() {}
   * }
   * frontend.inject("myapi", new MyAPI());
   * frontend.myapi.myfunction();
   *
   * @param {string} name the name at which to inject the object
   * @param {*} toInject the object to inject into the frontend
   */
  inject(name, toInject) {
    this.injected[name] = toInject;
    this[name] = toInject;
  }
}

export default Frontend;
