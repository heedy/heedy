var configPageMap = {};
var configRoutes = [];

class ConfigInjector {
    /**
     * Handle config
     * @param {*} frontend 
     */
    constructor(frontend) {
        this.store = frontend.store;
    }
    addPage(p) {
        configPageMap[p.path] = p;
    }
    $onInit() {
        Object.values(configPageMap).forEach((r) => {
            if (r.path.startsWith("/")) {
                r.path = r.path.substring(1, r.path.length);
            }
            configRoutes.push(r);
        });
        this.store.commit("setConfigRoutes", configRoutes);
    }
}

export {
    configRoutes
};
export default ConfigInjector;