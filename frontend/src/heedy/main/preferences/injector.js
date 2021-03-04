var preferencesPageMap = {};
var preferencesRoutes = [];

class PreferencesInjector {
    /**
     * Handle preferences
     * @param {*} frontend 
     */
    constructor(frontend) {
        this.store = frontend.store;
    }
    addPage(p) {
        preferencesPageMap[p.path] = p;
    }
    $onInit() {
        Object.values(preferencesPageMap).forEach((r) => {
            if (r.path.startsWith("/")) {
                r.path = r.path.substring(1, r.path.length);
            }
            preferencesRoutes.push(r);
        });
        this.store.commit("setPreferencesRoutes", preferencesRoutes);
    }
}

export {
    preferencesRoutes
};
export default PreferencesInjector;