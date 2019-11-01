var appRoutesMap = {};
var appRoutes = [];

class App {
    constructor(app) {
        this.store = app.store;

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
        if (app.info.user != null) {
            app.websocket.subscribe("app_create", {
                event: "app_create",
                user: app.info.user.username
            }, queryApp);
            app.websocket.subscribe("app_update", {
                event: "app_update",
                user: app.info.user.username
            }, queryApp);
            app.websocket.subscribe("app_delete", {
                event: "app_delete",
                user: app.info.user.username
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

    addRoute(r) {
        appRoutesMap[r.path] = r;
    }

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
export default App;