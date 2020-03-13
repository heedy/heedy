import Vue from "../dist/vue.mjs";
import vuexModule from "./main/vuex.js";
import Notification from "./main/notification.vue";
import NotificationsPage from "./main/notifications_page.vue";
import AppComponent from "./main/app_component.vue";
import ObjectComponent from "./main/object_component.vue";
import MenuIcon from "./main/menu_icon.vue";

function setup(frontend) {
    Vue.component("h-notification", Notification)

    if (frontend.info.user != null) {
        frontend.store.registerModule("notifications", vuexModule);

        frontend.apps.addComponent({
            key: "notifications",
            weight: 0.1,
            component: AppComponent
        });
        frontend.objects.addComponent({
            key: "notifications",
            weight: 0.1,
            component: ObjectComponent
        });

        frontend.addRoute({
            path: "/notifications",
            component: NotificationsPage
        });

        frontend.addMenuItem({
            key: "notifications",
            text: "Notifications",
            component: MenuIcon,
            route: "/notifications",
            location: "primary_bottom"
        });

        let notifier = (e) => {
            if (e.event.includes("delete")) {
                frontend.store.commit("deleteNotification", e.data);
            } else {
                frontend.store.commit("setNotification", e.data);
            }
        }

        let types = ["user", "app", "object"];
        let etypes = ["create", "update", "delete"]
        types.forEach((t) => etypes.forEach((et => {

            let etype = `${t}_notification_${et}`;
            frontend.websocket.subscribe(etype, {
                event: etype,
                user: frontend.info.user.username
            }, notifier);
        })));

        frontend.store.dispatch("readGlobalNotifications");
    }

}

export default setup;