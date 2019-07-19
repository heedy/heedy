
import Theme from "./main/theme.vue";

import PublicHome from "./main/public_home.vue";
import Login from "./main/login.vue";
import Logout from "./main/logout.vue";
import Settings from "./main/settings.vue";
import User from "./main/user.vue";

import vuexModule from "./main/statemanager.js";

function setup(app) {
    app.addVuexModule(vuexModule);
    
    app.setTheme(Theme);

    
    if (app.info.user!=null) {
        // Pages to set up when user is logged in
        app.addSecondaryMenuItem({
            key: "heedySettings",
            text: "Settings",
            icon: "mi:settings",
            route: "/settings"
        });
        

        app.addRoute({
            path: "/logout",
            component: Logout
        });
        app.addRoute({
            path: "/settings",
            component: Settings
        });

    } else {
        // Pages to set up for public site visitors
        app.addRoute({
            path: "/",
            component: PublicHome
        });
        app.addRoute({
            path: "/login",
            component: Login
        });

        app.addMenuItem({
            key: "heedyHome",
            text: "Home",
            icon: "mi:home",
            route: "/"
        });
    } 
    
    // Pages that are active in all situations

    app.addRoute({
        path: "/user/:username",
        component: User
    })
}

export default setup;
