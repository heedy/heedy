var appRoutesMap = {};
var appRoutes = [];

/**
 * @alias frontend.apps
 */
class AppInjector {
    /**
     * This portion of the API handles all pages relevant to apps.
     */
    constructor(frontend) {
        this.store = frontend.store;

        // queryApp is called each time an event happens involving a app.
        // It is used to keep the apps up-to-date
        let queryApp = (e) => {
            if (this.store.state.heedy.apps !== null) {
                this.store.dispatch("readApp_", {
                    id: e.app
                });
            }

        }


        // Subscribe to all app events, so that the app list
        // can be kept up-to-date
        if (frontend.info.user != null) {
            frontend.websocket.subscribe("app_create", {
                event: "app_create",
                user: frontend.info.user.username
            }, queryApp);
            frontend.websocket.subscribe("app_update", {
                event: "app_update",
                user: frontend.info.user.username
            }, queryApp);
            frontend.websocket.subscribe("app_delete", {
                event: "app_delete",
                user: frontend.info.user.username
            }, (e) => {
                if (this.store.state.heedy.apps !== null) {
                    // Instead of querying the deleted app, perform the delete explicitly
                    this.store.commit("setApp", {
                        id: e.app,
                        isNull: true
                    });
                }

            });
        }
    }

    /**
     * This function works in the same way as frontend.addRoute, but each path is relative to the app ID.
     * The components are passed a valid app object as the app prop.
     * @example
     * frontend.apps.addRoute({
     *  path: "myplugin/path", // This means /#/apps/:appid/myplugin/path
     *  component: MyComponent
     * });
     * 
     * @param {string} r.path The path, relative to /#/apps/:appid
     * @param {vue.Component} r.component Vue component object to show as the page at that route. 
     *          The component should have an `app` prop of type Object that is given the specific app.
     */
    addRoute(r) {
        appRoutesMap[r.path] = r;
    }

    /**
     * Add a component to display on each app's page (/#/apps/myappid)
     * @param {object} c Object containing component and display information
     * @param {string} c.key Key of the component, calling addComponent
     *          multiple times with the same key replaces the existing component
     *          with the new one. By default, heedy defines the "header" key, which
     *          contains the main card containing app icon and main options,
     *          and the "objects" key, which is the list of the app's objects. The
     *          notifications plugin adds a "notifications" component, which is only
     *          rendered when there are notifications for the app.
     * @param {float} c.weight the component's weight, with heavier components coming below
     *          lighter ones. The header has weight 0, and object list has weight 1. Notifications have weight 0.1.
     * @param {vue.Component} c.component The vue component object to display. Takes "app" object
     *          as a prop.
     * 
     * @example
     * frontend.apps.addComponent({
     *  key: "myComponentKey",
     *  weight: 2,
     *  component: MyComponent
     * });
     */
    addComponent(c) {
        this.store.commit("addAppComponent", c);
    }

    $onInit() {
        Object.values(appRoutesMap).reduce((_, r) => {
            if (r.path.startsWith("/")) {
                r.path = r.path.substring(1, r.path.length);
            }
            appRoutes.push(r);
            return null;
        }, null);
    }
}
export {
    appRoutes
}
export default AppInjector;