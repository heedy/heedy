import WorkerInjector from "./worker_injector.js";
import WebsocketInjector from "./websocket.js";

class Frontend {

    /**
     * Frontend is the main stuff
     */
    constructor(Vue, appinfo, store) {
        /**
         * The Vue instance. This is mainly to be used to register components and plugins.
         * 
         * @example
         * frontend.vue.use(MyVuePlugin);
         * frontend.vue.component("mycomponent",MyComponent);
         */
        this.vue = Vue;

        /**
         * The vuex store
         */
        this.store = store;

        /**
         * This property contains information passed in from the server:
         * @example
         * {
         *    "hi": "ho"
         * }
         */
        this.info = appinfo;

        /**
         * The vue component to use as the main theme for the frontend. The theme
         * renders the main menu, and holds the router that shows individual pages.
         * @example
         * frontend.theme = MyThemeComponent
         */
        this.theme = null;

        /**
         * The vue component to display when linked to a route that was not registered.
         * For example, the notFound component will be displayed when the path `#/blahblah`
         * is used.
         * @example
         * frontend.notFound = MyNotFoundComponent
         */
        this.notFound = null;

        /**
         * The worker is an instance of the Worker class
         */
        this.worker = new WorkerInjector(appinfo);
        this.websocket = new WebsocketInjector(this);


        this.injected = {};
        this.routes = {};




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

export default Frontend;