import Vue from "../dist.mjs";
import vuexModule from "./main/vuex.js";
import Notification from "./main/notification.vue";
import NotificationsPage from "./main/notifications_page.vue";
import ConnectionComponent from "./main/connection_component.vue";
import SourceComponent from "./main/source_component.vue";
import MenuIcon from "./main/menu_icon.vue";

function setup(app) {
    Vue.component("h-notification", Notification)

    if (app.info.user != null) {
        app.addVuexModule(vuexModule);

        app.connection.addComponent({
            key: "notifications",
            weight: 0.1,
            component: ConnectionComponent
        });
        app.source.addComponent({
            key: "notifications",
            weight: 0.1,
            component: SourceComponent
        });

        app.addRoute({
            path: "/notifications",
            component: NotificationsPage
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