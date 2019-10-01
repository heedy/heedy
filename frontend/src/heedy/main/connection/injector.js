var connectionRoutesMap = {};
var connectionRoutes = [];

class Connection {
    constructor(store) {
        this.store = store;
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