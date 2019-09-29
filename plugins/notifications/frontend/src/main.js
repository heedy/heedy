import vuexModule from "./main/vuex.js";
import Notifications from "./main/notifications.vue";
import MenuIcon from "./main/menu_icon.vue";

function setup(app) {
    if (app.info.user != null) {
        app.addVuexModule(vuexModule);

        app.addRoute({
            path: "/notifications",
            component: Notifications
        });

        app.addMenuItem({
            key: "notifications",
            text: "Notifications",
            component: MenuIcon,
            route: "/notifications",
            location: "primary_bottom"
        });
    }

}

export default setup;