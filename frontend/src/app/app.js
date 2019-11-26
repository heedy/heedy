import WorkerInjector from "./worker_injector.js";
import WebsocketInjector from "./websocket.js";


class App {
    constructor(Vue, appinfo, store) {
        // Allows registration of components
        this.vue = Vue;
        // Allows setting stuff in the store
        this.store = store;

        this.info = appinfo;

        this.theme = null;
        this.injected = {};
        this.routes = {};
        this.notFound = null;


        this.worker = new WorkerInjector(appinfo);
        this.websocket = new WebsocketInjector(this);
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

export default App;