var userRoutesMap = {};
var userRoutes = [];
/**
 * @alias frontend.users
 */
class UserInjector {
    /**
     * This portion of the API handles pages dealing with users.
     * @param {*} frontend 
     */
    constructor(frontend) {
        this.store = frontend.store;

        // queryUser is called on user update and delete events.
        // it checks if the store is keeping track of the user,
        // and tells it to update the user if it is.
        let queryUser = (e) => {
            let username = e.user;
            if (this.store.state.heedy.users[username] !== undefined) {
                this.store.dispatch("readUser_", {
                    username
                });
            }

        }

        // TODO: This needs to be fixed in the server: right now wildcard subscription
        // to users fails, but it should just subscribe to public users.
        // It should also succeed if not logged in (public users only)
        if (frontend.info.user != null) {
            frontend.websocket.subscribe("user_update", {
                event: "user_update",
                user: frontend.info.user.username //"*"
            }, queryUser);
            frontend.websocket.subscribe("user_delete", {
                event: "user_delete",
                user: frontend.info.user.username //"*"
            }, (e) => {
                let username = e.user;
                if (this.store.state.heedy.users[username] !== undefined) {
                    // On delete, set explicitly instead of re-querying
                    this.store.commit("setUser", {
                        username,
                        isNull: true
                    });
                }

            });
        }

    }
    /**
     * Adds a route to the user. Same as objects/apps addRoute. 
     * The resulting page will be rendered with prefix /#/users/myuser/.
     * 
     * Adding these paths can only be done during app initialization.
     * @param {object} r Object containing route information
     * @param {string} r.path The subpath at which to mount the route
     * @param {vue.Component} r.component The vue component to show at the path. 
     *          Takes user object as prop.
     * @example
     * frontend.users.addRoute({
     *  path: "mysubpath", // will be at /#/users/myuser/mysubpath
     *  component: MyComponent
     * });
     */
    addRoute(r) {
        userRoutesMap[r.path] = r;
    }

    /**
     * Add a component to display on each user's page (/#/users/myuser)
     * @param {object} c Object containing component and display information
     * @param {string} c.key Key of the component, calling addComponent
     *          multiple times with the same key replaces the existing component
     *          with the new one. By default, heedy defines the "header" key, which
     *          contains the main card containing user picture and descrition,
     *          and the "objects" key, which is the list of the user's objects
     * @param {float} c.weight the component's weight, with heavier components coming below
     *          lighter ones. The header has weight 0, and object list has weight 1.
     * @param {vue.Component} c.component The vue component object to display. Takes "user" object
     *          as a prop.
     * 
     * @example
     * frontend.users.addComponent({
     *  key: "myComponentKey",
     *  weight: 2,
     *  component: MyComponent
     * });
     */
    addComponent(c) {
        this.store.commit("addUserComponent", c);
    }

    /**
   * A function that given a user object, returns a map where each key is menu item key, and each value is
   * a menu item, and has icon, text, and action props.
   * @param {function(object)} mf 
   * 
   * @example
   * frontend.objects.addMenu((o)=> ({
   *  my_menu_item: {
   *    text: "My Menu Item",
   *    icon: "fas fa-code",
   *    path: `myplugin/${o.id}`
   *  }
   * }));
   */
    addMenu(mf) {
        this.store.commit("addUserMenu", mf)
    }

    $onInit() {
        Object.values(userRoutesMap).reduce((_, r) => {
            if (r.path.startsWith("/")) {
                r.path = r.path.substring(1, r.path.length);
            }
            userRoutes.push(r);
            return null;
        }, null);
    }
}

export {
    userRoutes
}
export default UserInjector;