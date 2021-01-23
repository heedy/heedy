import VueCodemirror from "../dist/codemirror.mjs";
import Draggable from "../dist/draggable.mjs";

import Theme from "./main/theme.vue";

import AboutPage from "./main/about.vue";
import Login from "./main/login.vue";
import Logout from "./main/logout.vue";

import SettingsPage from "./main/settings/index.vue";
import SettingsInjector, { settingsRoutes } from "./main/settings/injector.js";
import SettingsServer from "./main/settings/server.vue";
import SettingsUsers from "./main/settings/users.vue";
import SettingsPlugins from "./main/settings/plugins.vue";

import UserInjector, { userRoutes } from "./main/user/injector.js";
import UserRouter from "./main/user/router.vue";
import User from "./main/user/index.vue";
import UserHeader from "./main/user/header.vue";
import UserObjects from "./main/user/objects.vue";

import ObjectInjector, { objectRoutes } from "./main/object/injector.js";
import ObjectIndex from "./main/object/index.vue";
import ObjectUpdate from "./main/object/update.vue";
import ObjectRouter from "./main/object/router.vue";
import ObjectHeader from "./main/object/header_default.vue";
import ObjectBody from "./main/object/body_default.vue";
import ObjectList from "./main/object/list.vue";

import AppInjector, { appRoutes } from "./main/app/injector.js";
import AppRouter from "./main/app/router.vue";
import App from "./main/app/index.vue";
import AppHeader from "./main/app/header.vue";
import AppCreate from "./main/app/create.vue";
import AppUpdate from "./main/app/update.vue";
import AppObjects from "./main/app/objects.vue";
import AppSettings from "./main/app/settings.vue";

import Apps from "./main/apps.vue";

import vuexModule from "./main/vuex.js";
import registerCoreComponents, { NotFound } from "./main/components.js";

import moment from "../dist/moment.mjs";

function setup(frontend) {
  frontend.vue.use(VueCodemirror);
  frontend.vue.component("draggable", Draggable);

  frontend.theme = Theme;
  frontend.notFound = NotFound;

  // Add the current user to the vuex module
  if (frontend.info.user != null) {
    vuexModule.state.users[frontend.info.user.username] = {
      ...frontend.info.user,
      qtime: moment()
    };
  }
  frontend.store.registerModule("heedy", vuexModule);

  // Adds the components that are used throughout the UI
  registerCoreComponents(frontend.vue);
  frontend.vue.component("h-object-list", ObjectList);

  // Inject the user/app/object handlers into the frontend
  frontend.inject("users", new UserInjector(frontend));
  frontend.inject("apps", new AppInjector(frontend));
  frontend.inject("objects", new ObjectInjector(frontend));
  frontend.inject("settings", new SettingsInjector(frontend));

  frontend.users.addComponent({
    key: "header",
    weight: 0,
    component: UserHeader,
  });
  frontend.users.addComponent({
    key: "objects",
    weight: 1,
    component: UserObjects,
  });
  frontend.users.addRoute({
    path: "/",
    component: User,
  });

  if (frontend.info.user != null) {
    // Pages to set up when user is logged in

    frontend.addRoute({
      path: "/logout",
      component: Logout,
    });

    frontend.addRoute({
      path: "/apps",
      component: Apps,
    });
    frontend.addRoute({
      path: "/apps/:appid",
      props: true,
      component: AppRouter,
      children: appRoutes,
    });

    frontend.apps.addRoute({
      path: "",
      component: App,
    });

    frontend.apps.addRoute({
      path: "update",
      component: AppUpdate,
    });
    frontend.apps.addRoute({
      path: "settings",
      component: AppSettings,
    });

    // Add the default app UI
    frontend.apps.addComponent({
      key: "header",
      weight: 0,
      component: AppHeader,
    });
    frontend.apps.addComponent({
      key: "objects",
      weight: 1,
      component: AppObjects,
    });

    frontend.addRoute({
      path: "/create/app",
      component: AppCreate,
    });

    frontend.addRoute({
      path: "/",
      redirect: `/users/${frontend.info.user.username}`,
    });

    frontend.addMenuItem({
      key: "apps",
      text: "Apps",
      icon: "settings_input_component",
      route: "/apps",
      location: "secondary",
    });

    // Pages to show when the user is an admin
    if (frontend.info.admin) {
      frontend.addMenuItem({
        key: "heedySettings",
        text: "Settings",
        icon: "settings",
        route: "/settings/plugins",
        location: "secondary",
      });
      frontend.addRoute({
        path: "/settings",
        component: SettingsPage,
        children: settingsRoutes,
      });
      frontend.settings.addPage({
        path: "users",
        component: SettingsUsers,
        title: "Users",
      });
      frontend.settings.addPage({
        path: "server",
        component: SettingsServer,
        title: "Server",
      });
      frontend.settings.addPage({
        path: "plugins",
        component: SettingsPlugins,
        title: "Plugins",
      });
    }
  } else {
    // Pages to set up for public site visitors
    frontend.addRoute({
      path: "/about",
      component: AboutPage,
    });
    frontend.addRoute({
      path: "/login",
      component: Login,
    });

    frontend.addMenuItem({
      key: "about",
      text: "About",
      icon: "help_outline",
      route: "/about",
      location: "primary",
    });

    frontend.addRoute({
      path: "/",
      redirect: `/login`,
    });
  }

  // Pages that are active in all situations

  frontend.addRoute({
    path: "/users/:username",
    props: true,
    component: UserRouter,
    children: userRoutes,
  });

  frontend.addRoute({
    path: "/objects/:objectid",
    props: true,
    component: ObjectRouter,
    children: objectRoutes,
  });

  frontend.objects.addRoute({
    path: "/",
    component: ObjectIndex,
  });
  frontend.objects.addRoute({
    path: "/update",
    component: ObjectUpdate,
  });

  frontend.objects.addComponent({
    key: "header",
    weight: 0,
    component: ObjectHeader,
  });
  frontend.objects.addComponent({
    key: "body",
    weight: 5,
    component: ObjectBody,
  });
}

export default setup;
