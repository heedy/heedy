import SettingsPage from "./main/settings_page.vue";

function setup(app) {
    if (app.info.admin) {
        app.settings.addPage({
            path: "plugins",
            component: SettingsPage,
            title: "Plugins"
        });
    }
}

export default setup;