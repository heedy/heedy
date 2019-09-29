import vuexModule from "../statemanager.js";
var components = {};

var userRoutesMap = {};
var userRoutes = [];
class User {
    constructor(pluginName) {

    }

    addRoute(r) {
        userRoutesMap[r.path] = r;
    }

    addComponent(c) {
        components[c.key] = c;
    }

    static $onInit() {
        vuexModule.state.user_components = Object.values(components).sort((a, b) => a.weight - b.weight);

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