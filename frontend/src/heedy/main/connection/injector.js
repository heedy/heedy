import vuexModule from "../statemanager.js";
var components = {};

var connectionRoutesMap = {};
var connectionRoutes = [];

class Connection {
    constructor(pluginName) {

    }

    addRoute(r) {
        connectionRoutesMap[r.path] = r;
    }

    addComponent(c) {
        components[c.key] = c;
    }

    static $onInit() {
        vuexModule.state.connection_components = Object.values(components).sort((a, b) => a.weight - b.weight);

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