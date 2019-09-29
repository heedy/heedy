import Vue from "../dist.mjs";

import Theme from "./main/theme.vue";

import PublicHome from "./main/public_home.vue";
import Login from "./main/login.vue";
import Logout from "./main/logout.vue";
import Settings from "./main/settings.vue";


import UserInjector, {
    userRoutes
} from "./main/user/injector.js";
import UserRouter from "./main/user/router.vue";
import User from "./main/user/index.vue";
import UserHeader from "./main/user/header.vue";
import UserSources from "./main/user/sources.vue";

import SourceInjector, {
    sourceRoutes
} from "./main/source/injector.js";
import Source from "./main/source/index.vue";
import SourceRouter from "./main/source/router.vue";
import SourceHeader from "./main/source/header.vue";
import SourceList from "./main/source/list.vue";

import ConnectionInjector, {
    connectionRoutes
} from "./main/connection/injector.js";
import ConnectionRouter from "./main/connection/router.vue";
import Connection from "./main/connection/index.vue";
import ConnectionHeader from "./main/connection/header.vue";
import ConnectionCreate from "./main/connection/create.vue";
import ConnectionUpdate from "./main/connection/update.vue";
import ConnectionSources from "./main/connection/sources.vue";

import Connections from "./main/connections.vue";


import vuexModule from "./main/statemanager.js";
import registerCoreComponents from "./main/components.js";


function setup(app) {

    // Adds the components that are used throughout the UI
    registerCoreComponents(Vue);
    Vue.component("h-source-list", SourceList);

    // Inject the user/connection/source handlers into the app
    app.inject("user", UserInjector);
    app.inject("connection", ConnectionInjector);
    app.inject("source", SourceInjector);

    // Add the current user to the cache
    if (app.info.user != null) {
        vuexModule.state.users[app.info.user.name] = app.info.user;
    }
    app.addVuexModule(vuexModule);

    app.setTheme(Theme);

    app.user.addComponent({
        key: "header",
        weight: 0,
        component: UserHeader
    });
    app.user.addComponent({
        key: "sources",
        weight: 1,
        component: UserSources
    });
    app.user.addRoute({
        path: "/",
        component: User
    });


    if (app.info.user != null) {
        // Pages to set up when user is logged in
        app.addMenuItem({
            key: "heedySettings",
            text: "Settings",
            icon: "settings",
            route: "/settings",
            location: "secondary"
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
            props: true,
            component: ConnectionRouter,
            children: connectionRoutes
        });

        app.connection.addRoute({
            path: "",
            component: Connection
        })

        app.connection.addRoute({
            path: "update",
            component: ConnectionUpdate
        })


        // Add the default connection UI
        app.connection.addComponent({
            key: "header",
            weight: 0,
            component: ConnectionHeader
        });
        app.connection.addComponent({
            key: "sources",
            weight: 1,
            component: ConnectionSources
        });




        app.addRoute({
            path: "/create/connection",
            component: ConnectionCreate
        });

        app.addRoute({
            path: "/",
            redirect: `/users/${app.info.user.username}`
        });

        app.addMenuItem({
            key: "connections",
            text: "Connections",
            icon: "settings_input_component",
            route: "/connections",
            location: "primary",
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
            route: "/",
            location: "primary"
        });
    }

    // Pages that are active in all situations

    app.addRoute({
        path: "/users/:username",
        props: true,
        component: UserRouter,
        children: userRoutes
    });

    app.addRoute({
        path: "/sources/:sourceid",
        props: true,
        component: SourceRouter,
        children: sourceRoutes
    });

    app.source.addRoute({
        path: "/",
        component: Source
    })

    app.source.addComponent({
        key: "header",
        weight: 0,
        component: SourceHeader
    })

}

export default setup;