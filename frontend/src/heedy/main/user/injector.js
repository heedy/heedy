var userRoutesMap = {};
var userRoutes = [];
class User {
    constructor(app) {
        this.store = app.store;

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
        if (app.info.user != null) {
            app.websocket.subscribe("user_update", {
                event: "user_update",
                user: app.info.user.username //"*"
            }, queryUser);
            app.websocket.subscribe("user_delete", {
                event: "user_delete",
                user: app.info.user.username //"*"
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
export default User;