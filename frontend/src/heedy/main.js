
import Theme from "./main/theme.vue";

import PublicHome from "./main/public_home.vue";
import Login from "./main/login.vue";
import Logout from "./main/logout.vue";
import Settings from "./main/settings.vue";
import User from "./main/user.vue";
import Source from "./main/source.vue";
import Connections from "./main/connections.vue";
import Connection from "./main/connection.vue";
import CreateConnection from "./main/create_connection.vue";
import SourceRouter from "./main/source_router.vue";

import vuexModule from "./main/statemanager.js";
import SourceInjector, {sourceTypeRouter} from "./main/sourceInjector.js";



function setup(app) {
    // Inject the source handler to the app
    app.inject("source",SourceInjector);

    app.addVuexModule(vuexModule);
    
    app.setTheme(Theme);

    
    if (app.info.user!=null) {
        // Pages to set up when user is logged in
        app.addSecondaryMenuItem({
            key: "heedySettings",
            text: "Settings",
            icon: "settings",
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
        app.addRoute({
            path: "/connections",
            component: Connections
        });
        app.addRoute({
            path: "/connections/:connectionid",
            component: Connection
        });
        app.addRoute({
            path: "/create/connection",
            component: CreateConnection
        });

        app.addRoute({
            path: "/",
            redirect: `/user/${app.info.user.username}`
        });

        app.addMenuItem({
            key: "connections",
            text: "Connections",
            icon: "settings_input_component",
            route: "/connections"
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
            icon: "home",
            route: "/"
        });
    } 
    
    // Pages that are active in all situations

    app.addRoute({
        path: "/users/:username",
        props: true,
        component: User
    });

    app.addRoute({
        path: "/sources/:sourceid",
        props: true,
        component: SourceRouter,
        // The children are initialized by the injector.
        children: sourceTypeRouter
    });

    // Add the root source router
    sourceTypeRouter.push({
        path: '',
        props: true,
        component: Source
    });
    
}

export default setup;
