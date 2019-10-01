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
        app.store.registerModule("notifications", vuexModule);

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

        let notifier = (e) => {
            if (e.event.includes("delete")) {
                app.store.commit("deleteNotification", e.data);
            } else {
                app.store.commit("setNotification", e.data);
            }
        }

        let types = ["user", "connection", "source"];
        let etypes = ["create", "update", "delete"]
        types.forEach((t) => etypes.forEach((et => {

            let etype = `${t}_notification_${et}`;
            app.events.subscribe(etype, {
                event: etype,
                user: app.info.user.username
            }, notifier);
        })));
    }

}

export default setup;