var settingsPageMap = {};
var settingsRoutes = [];

/**
 * @alias frontend.settings
 */
class SettingsInjector {
    /**
     * Handle settings
     * @param {*} frontend 
     */
    constructor(frontend) {
        this.store = frontend.store;
    }
    addPage(p) {
        settingsPageMap[p.path] = p;
    }

    /**
     * Use a custom component for the settings page for the given plugin's user settings.
     * @param {string} plugin The name of the plugin for which to show a custom editor component
     * @param {vue.Component} component The component, with props schema,value,plugin, which emits an update event with the new settings.
     */
    setUserSettingsComponent(plugin, component) {
        this.store.commit("setUserSettingsComponent", { plugin, component });
    }
    $onInit() {
        Object.values(settingsPageMap).forEach((r) => {
            if (r.path.startsWith("/")) {
                r.path = r.path.substring(1, r.path.length);
            }
            settingsRoutes.push(r);
        });
        this.store.commit("setSettingsRoutes", settingsRoutes);
    }
}

export {
    settingsRoutes
};
export default SettingsInjector;