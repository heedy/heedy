import VueCodemirror from "../dist/codemirror.mjs";
import Draggable from "../dist/draggable.mjs";

import Theme from "./main/theme.vue";

import AboutPage from "./main/about.vue";
import Login from "./main/login.vue";
import Logout from "./main/logout.vue";

import ConfigPage from "./main/config/index.vue";
import ConfigInjector, { configRoutes } from "./main/config/injector.js";
import ConfigConfiguration from "./main/config/configuration.vue";
import ConfigUsers from "./main/config/users.vue";
import ConfigPlugins from "./main/config/plugins.vue";

import SettingsPage from "./main/settings/index.vue";
import SettingsInjector, { settingsRoutes } from "./main/settings/injector.js";
import SettingsUserEdit from "./main/settings/useredit.vue";
import SettingsUserSettings from "./main/settings/usersettings.vue";
import SettingsSessions from "./main/settings/usersessions.vue";

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
import AppToolbarItem from "./main/app/toolbar_item.vue";
import AppAccessTokenToolbarItem from "./main/app/toolbar_access_token_item.vue";
import AppSettingsToolbarItem from "./main/app/toolbar_settings_item.vue";

import Apps from "./main/apps.vue";

import vuexModule from "./main/vuex.js";
import registerCoreComponents, { NotFound } from "./main/components.js";

function setup(frontend) {
  frontend.vue.use(VueCodemirror);
  frontend.vue.component("draggable", Draggable);

  frontend.theme = Theme;
  frontend.notFound = NotFound;

  // Add the current user to the vuex module
  if (frontend.info.user != null) {
    vuexModule.state.users[frontend.info.user.username] = frontend.info.user;
    vuexModule.state.users_qtime[frontend.info.user.username] = new Date();
  }
  frontend.store.registerModule("heedy", vuexModule);

  // Adds the components that are used throughout the UI
  registerCoreComponents(frontend.vue);
  frontend.vue.component("h-object-list", ObjectList);
  // Add form elements for the json schema component
  frontend.inject("addSchemaFormElement",(k,c) => frontend.store.commit("addSchemaFormElement",{key:k,component:c}));

  // Inject the user/app/object handlers into the frontend
  frontend.inject("users", new UserInjector(frontend));
  frontend.inject("apps", new AppInjector(frontend));
  frontend.inject("objects", new ObjectInjector(frontend));
  frontend.inject("config", new ConfigInjector(frontend));
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

    // Set up websocket listening for preference updates
    frontend.websocket.subscribe("user_settings_update", {
      event: "user_settings_update",
      user: frontend.info.user.username //"*"
    }, e => frontend.store.dispatch("ReadUserPluginSettings", { plugin: e.plugin }))

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

    frontend.addMenuItem({
      key: "heedySettings",
      text: "User Settings",
      icon: "fas fa-user-cog",
      route: "/settings/plugins",
      location: "secondary",
    });
    frontend.addRoute({
      path: "/settings",
      component: SettingsPage,
      children: settingsRoutes,
    });
    frontend.settings.addPage({
      path: "user",
      component: SettingsUserEdit,
      title: "My Account"
    });
    frontend.settings.addPage({
      path: "plugins",
      component: SettingsUserSettings,
      title: "Settings"
    });
    frontend.settings.addPage({
      path: "sessions",
      component: SettingsSessions,
      title: "Sessions"
    });

    // Pages to show when the user is an admin
    if (frontend.info.admin) {
      frontend.addMenuItem({
        key: "heedyConfig",
        text: "Server Config",
        icon: "settings",
        route: "/config/plugins",
        location: "secondary",
      });
      frontend.addRoute({
        path: "/config",
        component: ConfigPage,
        children: configRoutes,
      });
      frontend.config.addPage({
        path: "users",
        component: ConfigUsers,
        title: "Users",
      });
      frontend.config.addPage({
        path: "configuration",
        component: ConfigConfiguration,
        title: "config",
      });
      frontend.config.addPage({
        path: "plugins",
        component: ConfigPlugins,
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


  // Toolbars to show in User/object/app headers
  frontend.objects.addMenu((o) => {
    let m = {};
    let access = o.access.split(" ");
    if (access.includes('*') || access.includes('write')) {
      m["edit"] = {
        icon: "edit",
        text: "Edit",
        to: `/objects/${o.id}/update`,
        toolbar: true
      };
    }
    if (o.app != null) {
      m["app"] = {
        icon: "code",
        text: "Go To App",
        to: `/apps/${o.app}`,
        toolbar_component: AppToolbarItem,
        menu_component: AppToolbarItem,
        toolbar_props: { appid: o.app },
        menu_props: { appid: o.app, isList: true },
        toolbar: true,
        weight: 1
      };
    }
    return m;
  });

  frontend.apps.addMenu((app) => {
    let m = {
      edit: {
        icon: "edit",
        text: "Edit App",
        to: `/apps/${app.id}/update`,
        toolbar: true
      }
    };
    if (app.access_token === undefined || app.access_token != '') {
      m["access_token"] = {
        toolbar: true,
        toolbar_component: AppAccessTokenToolbarItem,
        menu_component: AppAccessTokenToolbarItem,
        toolbar_props: { appid: app.id },
        menu_props: { appid: app.id, isList: true },
        weight: -0.5
      }
    }
    if (Object.keys(app.settings_schema).length > 0) {
      m["settings"] = {
        icon: "fas fa-cog",
        text: "App Settings",
        to: `/apps/${app.id}/settings`,
        toolbar: true,
        toolbar_component: AppSettingsToolbarItem,
        menu_component: AppSettingsToolbarItem,
        toolbar_props: { appid: app.id },
        menu_props: { appid: app.id, isList: true },
        weight: 1
      }
    }
    return m;
  });


}

export default setup;
