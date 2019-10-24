var connectionRoutesMap = {};
var connectionRoutes = [];

class Connection {
    constructor(app) {
        this.store = app.store;

        // queryConnection is called each time an event happens involving a connection.
        // It is used to keep the connections up-to-date
        let queryConnection = (e) => {
            if (this.store.state.heedy.connections !== null) {
                this.store.dispatch("readConnection_", {
                    id: e.connection
                });
            }

        }


        // Subscribe to all connection events, so that the connection list
        // can be kept up-to-date
        if (app.info.user != null) {
            app.websocket.subscribe("connection_create", {
                event: "connection_create",
                user: app.info.user.username
            }, queryConnection);
            app.websocket.subscribe("connection_update", {
                event: "connection_update",
                user: app.info.user.username
            }, queryConnection);
            app.websocket.subscribe("connection_delete", {
                event: "connection_delete",
                user: app.info.user.username
            }, (e) => {
                if (this.store.state.heedy.connections !== null) {
                    // Instead of querying the deleted connection, perform the delete explicitly
                    this.store.commit("setConnection", {
                        id: e.connection,
                        isNull: true
                    });
                }

            });
        }
    }

    addRoute(r) {
        connectionRoutesMap[r.path] = r;
    }

    addComponent(c) {
        this.store.commit("addConnectionComponent", c);
    }

    $onInit() {
        Object.values(connectionRoutesMap).reduce((_, r) => {
            if (r.path.startsWith("/")) {
                r.path = r.path.substring(1, r.path.length);
            }
            connectionRoutes.push(r);
            return null;
        }, null);
    }
}
export {
    connectionRoutes
}
export default Connection;