import vuexModule from "./statemanager.js";

var sourceRoutes = {};

var sourceTypeRouter = [];

var typeComponents = {};


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

    typeComponent(t,c) {
        typeComponents[t] = c;
    }

    /**
     * Adds a route to the given source type. The route
     * automatically takes /source/:sourceid/{r.path}
     * @param {*} r 
     */
    addRoute(r) {
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

export {sourceTypeRouter, typeComponents}
export default Source;