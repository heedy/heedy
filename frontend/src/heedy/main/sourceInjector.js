import vuexModule from "./statemanager.js";

var sourceRoutes = {};

var sourceTypeRouter = [];


class Source {
    constructor(pluginName) {
    }

    /**
     * Identical to a menu item, it is displayed in a special source creation menu
     * @param {*} c The creator to add
     */
    addCreator(c) {
        vuexModule.state.sourceCreators.push(c);
    }

    typePath(t,p) {
        vuexModule.state.typePaths[t] = p;
    }

    /**
     * Adds a route to the given source type. The route
     * automatically takes /source/:sourceid/{t}/ as the root path
     * @param {*} t 
     * @param {*} r 
     */
    addRoute(t,r) {
        r.path = t + r.path;
        sourceRoutes[r.path] = r;
    }

    static $onInit() {
        // Need to set the sourceTypeRouter with the right values:
        for (let key in sourceRoutes) {
            // skip loop if the property is from prototype
            if (!sourceRoutes.hasOwnProperty(key)) continue;
            sourceTypeRouter.push(sourceRoutes[key]);
        }
    }
}

export {sourceTypeRouter}
export default Source;