import VueCodemirror from "../dist/codemirror.mjs";
import Draggable from "../dist/draggable.mjs";


import Theme from "./main/theme.vue";

import PublicHome from "./main/public_home.vue";
import Login from "./main/login.vue";
import Logout from "./main/logout.vue";

import SettingsPage from "./main/settings/index.vue";
import SettingsInjector, {
    settingsRoutes
} from "./main/settings/injector.js";
import SettingsServer from "./main/settings/server.vue";
import SettingsUsers from "./main/settings/users.vue";
import SettingsPlugins from "./main/settings/plugins.vue";


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

import AppInjector, {
    appRoutes
} from "./main/app/injector.js";
import AppRouter from "./main/app/router.vue";
import App from "./main/app/index.vue";
import AppHeader from "./main/app/header.vue";
import AppCreate from "./main/app/create.vue";
import AppUpdate from "./main/app/update.vue";
import AppSources from "./main/app/sources.vue";
import AppSettings from "./main/app/settings.vue";

import Apps from "./main/apps.vue";


import vuexModule from "./main/vuex.js";
import registerCoreComponents from "./main/components.js";

function setup(app) {
    app.vue.use(VueCodemirror);
    app.vue.component('draggable', Draggable);

    app.theme = Theme;

    // Add the current user to the vuex module
    if (app.info.user != null) {
        vuexModule.state.users[app.info.user.username] = app.info.user;
    }
    app.store.registerModule("heedy", vuexModule);

    // Adds the components that are used throughout the UI
    registerCoreComponents(app.vue);
    app.vue.component("h-source-list", SourceList);

    // Inject the user/app/source handlers into the app
    app.inject("user", new UserInjector(app));
    app.inject("app", new AppInjector(app));
    app.inject("source", new SourceInjector(app));
    app.inject("settings", new SettingsInjector(app));



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
        if (app.info.admin) {
            app.addMenuItem({
                key: "heedySettings",
                text: "Settings",
                icon: "settings",
                route: "/settings/plugins",
                location: "secondary"
            });
            app.addRoute({
                path: "/settings",
                component: SettingsPage,
                children: settingsRoutes
            });
            app.settings.addPage({
                path: "users",
                component: SettingsUsers,
                title: "Users"
            });
            app.settings.addPage({
                path: "server",
                component: SettingsServer,
                title: "Server"
            });
            app.settings.addPage({
                path: "plugins",
                component: SettingsPlugins,
                title: "Plugins"
            });
        }



        app.addRoute({
            path: "/logout",
            component: Logout
        });


        app.addRoute({
            path: "/apps",
            component: Apps
        });
        app.addRoute({
            path: "/apps/:appid",
            props: true,
            component: AppRouter,
            children: appRoutes
        });

        app.app.addRoute({
            path: "",
            component: App
        });

        app.app.addRoute({
            path: "update",
            component: AppUpdate
        });
        app.app.addRoute({
            path: "settings",
            component: AppSettings
        });


        // Add the default app UI
        app.app.addComponent({
            key: "header",
            weight: 0,
            component: AppHeader
        });
        app.app.addComponent({
            key: "sources",
            weight: 1,
            component: AppSources
        });




        app.addRoute({
            path: "/create/app",
            component: AppCreate
        });

        app.addRoute({
            path: "/",
            redirect: `/users/${app.info.user.username}`
        });

        app.addMenuItem({
            key: "apps",
            text: "Apps",
            icon: "settings_input_component",
            route: "/apps",
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