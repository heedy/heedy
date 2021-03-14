import api from "../util.mjs";
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
   * @param {boolean} json - whether data should be sent as standard POST url encoded or as json
   *
   * @returns {object} an object with two fields, response and data. response gives a fetch query response object, and data contains the response content decoded from json.
   */
  async rest(method, uri, data = null, params = null, json = true) {
    return await api(method, uri, data, params, json);
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
   * @param {object} m Menu item to add. It is given an object
   *        with items "key", which is a unique ID, text, the text to display,
   *        icon, the icon to show, and route, which is the route to navigate to.
   *        Optionally also has a "location" attribute which hints at where the user
   *        might want the menu (primary,secondary,spaced_primary).
   *        Can also have "component" which is a vue component to display instead of icon.
   *        Be aware that the component must have a "state" prop, where it is told how to behave
   *        i.e. whether the menu is small, on bottom, etc.
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
