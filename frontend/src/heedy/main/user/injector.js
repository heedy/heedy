var userRoutesMap = {};
var userRoutes = [];
class UserInjector {
    /**
     * Handle users
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

    addRoute(r) {
        userRoutesMap[r.path] = r;
    }

    addComponent(c) {
        this.store.commit("addUserComponent", c);
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