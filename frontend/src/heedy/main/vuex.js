import Vue from "../../dist/vue.mjs";
import api, {deepEqual} from "../../util.mjs";

import {isBefore,isAfter,subSeconds,addSeconds} from "../../dist/date-fns.mjs";

export default {
  state: {
    alert: {
      value: false,
      text: "",
      type: "",
    },
    users: {},
    users_qtime: {},

    // Components to show for a user
    user_components: [],

    // The current user's apps & time when they were queried
    apps: null,
    apps_qtime: null,

    // The map of available plugin apps
    plugin_apps: null,

    // Components to show in the app
    app_components: [],

    // A map of objects and the time they were queried.
    // qtime holds an array of callbacks for queries in progress
    // and the timestamp of queries already completed.
    objects: {},
    objects_qtime: {},
    objectMenu: [],

    // Components to show for a object
    object_components: [],
    // Object types hold the app customization for each object type
    object_types: {},

    // a map keyed by username, where each element is a map of ids to null
    userObjects: {},
    userObjects_qtime: {},
    userMenu: [],

    // a map keyed by app id, where each element is a map of ids to null
    appObjects: {},
    appObjects_qtime: {},
    appMenu: [],

    // The following are initialized by the objectInjector
    objectCreators: [],

    // Subpaths for each object type
    typePaths: {},

    // The map of app scopes along with their descriptions
    appScope: null,

    settings_routes: [],
    user_settings_custom_components: {},

    config_routes: [],
    updates: {
      heedy: false,
      plugins: [],
      config: false,
      options: null,
    },

    user_settings_schema: null,

    // Custom elements to show in json schema form
    schema_form_elements: {},
  },
  mutations: {
    addSchemaFormElement(state,{key,component}) {
      Vue.set(state.schema_form_elements,key,component);
    },
    setConfigRoutes(state, v) {
      state.config_routes = v;
    },
    setSettingsRoutes(state, v) {
      state.settings_routes = v;
    },
    setUserSettingsSchema(state, v) {
      state.user_settings_schema = v;
    },
    addAppComponent(state, v) {
      state.app_components.push(v);
    },
    addUserComponent(state, v) {
      state.user_components.push(v);
    },
    addUserMenu(state, v) {
      state.userMenu.push(v);
    },
    addAppMenu(state, v) {
      state.appMenu.push(v);
    },
    addObjectMenu(state, v) {
      state.objectMenu.push(v);
    },
    setObjectType(state, v) {
      if (state.object_types[v.type] !== undefined) {
        v = {
          ...state.object_types[v.type],
          ...v,
        };
      }
      state.object_types[v.type] = v;
    },
    addObjectComponent(state, v) {
      state.object_components.push(v);
    },
    addObjectCreator(state, c) {
      state.objectCreators.push(c);
    },
    alert(state, v) {
      state.alert = {
        value: true,
        type: "",
        text: "",
        ...v,
      };
    },
    setUser(state, v) {
      Vue.set(state.users_qtime, v.username, new Date());
      if (v.isNull !== undefined) {
        if (state.userObjects[v.username] !== undefined) {
          Vue.delete(state.userObjects, v.username);
        }
        Vue.set(state.users, v.username, null);
        return;
      }
      Vue.set(state.users, v.username, v);
    },
    setApp(state, v) {
      if (state.apps == null) {
        state.apps = {};
      }
      if (v.isNull !== undefined) {
        if (state.appObjects[v.id] !== undefined) {
          Vue.delete(state.appObjects, v.id);
        }
        if (state.apps[v.id] !== undefined) {
          Vue.delete(state.apps, v.id);
        }
        return;
      }
      Vue.set(state.apps, v.id, {
        qtime: new Date(),
        ...v,
      });
    },
    setApps(state, v) {
      let qtime = new Date();
      Object.keys(v).forEach((k) => {
        v[k] = {
          qtime,
          ...v[k],
        };
      });
      Object.keys(state.appObjects).forEach((k) => {
        if (v[k] === undefined) {
          Vue.delete(state.appObjects, k);
        }
      });
      state.apps = v;

    },
    setAppsQTime(state, t) {
      state.apps_qtime = t;
    },
    setObject(state, v) {
      // First check if the object has existing value
      let curs = state.objects[v.id] || null;
      // Get the callbacks
      let callbacks =
        state.objects_qtime[v.id] === undefined ||
          !Array.isArray(state.objects_qtime[v.id])
          ? []
          : state.objects_qtime[v.id];
      Vue.set(state.objects_qtime, v.id, new Date());

      if (deepEqual(curs, v)) {
        callbacks.forEach((c) => c());
        return;
      }

      if (v.isNull !== undefined) {
        // The object is to be deleted - make sure to take care of all places it could be
        if (curs !== null) {
          if (curs.app !== null) {
            if (state.appObjects[curs.app] !== undefined) {
              Vue.delete(state.appObjects[curs.app], curs.id);
            }
          } else if (state.userObjects[curs.owner] !== undefined) {
            Vue.delete(state.userObjects[curs.owner], curs.id);
          }
        }
        Vue.set(state.objects, v.id, null);

        callbacks.forEach((c) => c());

        return;
      }
      // Set the object
      Vue.set(state.objects, v.id, v);

      // Delete from lists where changed
      if (curs != null) {
        if (v.app != curs.app) {
          if (state.appObjects[curs.app] !== undefined) {
            Vue.delete(state.appObjects[curs.app], curs.id);
          }
        }
        if (v.owner != curs.owner) {
          if (state.userObjects[curs.owner] !== undefined) {
            Vue.delete(state.userObjects[curs.owner], curs.id);
          }
        }
      }

      // Make sure to set it in the appropriate lists
      if (v.app != null && state.appObjects[v.app] !== undefined) {
        Vue.set(state.appObjects[v.app], v.id, null);
      }
      if (state.userObjects[v.owner] !== undefined) {
        Vue.set(state.userObjects[v.owner], v.id, null);
      }

      callbacks.forEach((c) => c());
    },
    addObjectQTimeCallback(state, v) {
      if (
        state.objects_qtime[v.id] === undefined ||
        !Array.isArray(state.objects_qtime[v.id])
      ) {
        state.objects_qtime[v.id] = [];
      }
      if (v.callback !== undefined) {
        state.objects_qtime[v.id].push(v.callback);
      }
    },
    setUserObjects(state, v) {
      let srcidmap = {};
      let qtime =new Date();
      v.objects.forEach((s) => {
        srcidmap[s.id] = null;
      });
      Vue.set(state.userObjects, v.user, srcidmap);
      Vue.set(state.userObjects_qtime, v.user, qtime);
    },
    setUserObjectsQTime(state, uname) {
      Vue.set(state.userObjects_qtime, uname, new Date());
    },
    setAppObjects(state, v) {
      let srcidmap = {};
      let qtime = new Date();
      v.objects.forEach((s) => {
        srcidmap[s.id] = null;
      });
      Vue.set(state.appObjects, v.id, srcidmap);
      Vue.set(state.appObjects_qtime, v.id, qtime);
    },
    setAppScope(state, v) {
      state.appScope = v;
    },
    setUpdates(state, v) {
      state.updates = v;
    },
    setPluginApps(state, v) {
      state.plugin_apps = v;
    },
    setUserSettingsComponent(state, v) {
      state.user_settings_custom_components[v.plugin] = v.component;
    },
  },
  actions: {
    errnotify({ commit }, v) {
      // Notifies of an error
      if (v.hasOwnProperty("error")) {
        // Only notify if it is an actual error
        commit("alert", {
          type: "error",
          text: v.error_description,
        });
      }
    },
    // This function performs a query on the user, ignoring websocket
    readUser_: async function ({ commit, rootState }, q) {
      let username = q.username;
      console.vlog("Reading user", username);
      let res = await api("GET", `api/users/${encodeURIComponent(username)}`, {
        icon: true,
      });
      if (!res.response.ok) {
        // If the error is 404, set the user to null
        if (res.response.status == 400 || res.response.status == 403) {
          // TODO: 404 should be returned
          commit("setUser", {
            username: username,
            isNull: true,
          });
        } else {
          commit("alert", {
            type: "error",
            text: res.data.error_description,
          });
        }
      } else {
        if (
          rootState.app.info.user != null &&
          rootState.app.info.user.username == username
        ) {
          commit("updateLoggedInUser", res.data);
        }
        commit("setUser", res.data);
      }
      if (q.hasOwnProperty("callback")) {
        q.callback();
      }
    },
    readApp_: async function ({ commit }, q) {
      console.vlog("Reading app", q.id);
      let res = await api("GET", `api/apps/${encodeURIComponent(q.id)}`, {
        icon: true,
      });
      if (!res.response.ok) {
        if (res.response.status == 400 || res.response.status == 403) {
          // TODO: 404 should be returned
          commit("setApp", {
            id: q.id,
            isNull: true,
          });
        } else {
          commit("alert", {
            type: "error",
            text: res.data.error_description,
          });
        }
      } else {
        commit("setApp", res.data);
      }

      if (q.hasOwnProperty("callback")) {
        q.callback();
      }
    },
    readObject_: async function ({ commit, state }, q) {
      if (
        state.objects_qtime[q.id] !== undefined &&
        Array.isArray(state.objects_qtime[q.id])
      ) {
        console.vlog(`waiting for object ${q.id}`);
        if (q.callback !== undefined) {
          commit("addObjectQTimeCallback", {
            id: q.id,
            callback: q.callback,
          });
        }
        return;
      }

      // Set up the query waiting array
      commit("addObjectQTimeCallback", {
        id: q.id,
      });

      console.vlog("Reading object", q.id);
      let res = await api("GET", `api/objects/${encodeURIComponent(q.id)}`, {
        icon: true,
      });
      if (!res.response.ok) {
        if (res.response.status == 400 || res.response.status == 403) {
          // TODO: 404 should be returned
          commit("setObject", {
            id: q.id,
            isNull: true,
          });
        } else {
          commit("alert", {
            type: "error",
            text: res.data.error_description,
          });
        }
      } else {
        commit("setObject", res.data);
      }

      if (q.hasOwnProperty("callback")) {
        q.callback();
      }
    },

    readUser({ state, rootState, dispatch }, q) {
      let username = q.username;
      if (
        state.users[username] !== undefined &&
        state.users[username] != null
      ) {
        // If the user was queried up to 1 second before websocket became active,
        // or was queried less than a second ago, let's just leave it. This avoids
        // an unnecessary query to read user on app startup
        let cmptime = addSeconds(state.users_qtime[username],1);
        if (
          rootState.app.websocket != null &&
          isBefore(rootState.app.websocket,cmptime) || isBefore(new Date(),cmptime)
        ) {
          console.vlog(`Not querying ${username} - websocket active or just queried`);
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
          return;
        }
      }
      dispatch("readUser_", q);
    },
    readApp: async function ({ state, rootState, dispatch }, q) {
      if (state.apps == null) {
        dispatch("listApps", q);
        return;
      }
      if (state.apps[q.id] !== undefined) {

        if (
          rootState.app.websocket != null &&
          isBefore(rootState.app.websocket,state.apps[q.id].qtime)
        ) {
          console.vlog(`Not querying ${q.id} - websocket active`);
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
          return;
        }
        if (
          state.apps_qtime !== null &&
          isAfter(state.apps_qtime,
            subSeconds(new Date(),1)
          )
        ) {
          console.vlog(
            "Not re-reading apps - they were just queried!"
          );
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
          return;
        }
      }
      dispatch("readApp_", q);
    },
    readObject: async function ({ state, rootState, dispatch }, q) {
      if (state.objects[q.id] !== undefined && state.objects[q.id] !== null) {
        if (
          rootState.app.websocket != null &&
          isBefore(rootState.app.websocket,state.objects_qtime[q.id])
        ) {
          console.vlog(`Not querying ${q.id} - websocket active`);
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
          return;
        }
      }
      dispatch("readObject_", q);
    },
    readUserObjects: async function ({ commit, state, rootState }, q) {
      // Only if they are not being kept up-to-date by the websocket
      if (state.userObjects_qtime[q.username] !== undefined) {
        if (
          rootState.app.websocket !== null &&
          isBefore(rootState.app.websocket,state.userObjects_qtime[q.username])
        ) {
          console.vlog(`Not reading ${q.username} objects - websocket active`);
          return;
        }
        // Check if we JUST queried less than a second ago
        if (
          isAfter(state.userObjects_qtime[q.username],
            subSeconds(new Date(),1)
          )
        ) {
          console.vlog(
            `Not re-reading ${q.username} objects - they were just queried!`
          );
          return;
        }
      }
      commit("setUserObjectsQTime", q.username);
      console.vlog("Reading objects for user", q.username);
      let query = {
        owner: q.username,
        icon: true,
      };

      let res = await api("GET", `api/objects`, query);
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description,
        });
      } else {
        res.data.forEach((obj) => commit("setObject", obj));
        commit("setUserObjects", {
          user: q.username,
          objects: res.data,
        });
      }

      if (q.hasOwnProperty("callback")) {
        q.callback();
      }
    },
    readAppObjects: async function ({ commit, state, rootState }, q) {
      // Only if they are not being kept up-to-date by the websocket
      if (
        state.appObjects[q.id] !== undefined &&
        rootState.app.websocket !== null &&
        isBefore(rootState.app.websocket,state.appObjects_qtime[q.id])
      ) {
        console.vlog(`Not reading ${q.id} objects - websocket active`);
        return;
      }
      console.vlog("Reading objects for app", q.id);
      let query = {
        app: q.id,
        icon: true,
      };

      let res = await api("GET", `api/objects`, query);
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description,
        });
      } else {
        res.data.forEach((obj) => commit("setObject", obj));
        commit("setAppObjects", {
          id: q.id,
          objects: res.data,
        });
      }

      if (q.hasOwnProperty("callback")) {
        q.callback();
      }
    },
    getAppScope: async function ({ commit }) {
      console.vlog("Loading available app scopes");
      let res = await api("GET", "api/server/scope");
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description,
        });
      } else {
        commit("setAppScope", res.data);
      }
    },
    listApps: async function ({ commit, state, rootState }, q) {
      // Only list apps if they are not being kept up-to-date by the websocket
      if (
        state.apps !== null &&
        rootState.app.websocket !== null &&
        isBefore(rootState.app.websocket,state.apps_qtime)
      ) {
        console.vlog("Not listing apps - websocket active");
        return;
      }
      if (
        state.apps_qtime !== null &&
        isAfter(state.apps_qtime,
         subSeconds(new Date(),1)
        )
      ) {
        console.vlog(
          `Not re-reading apps - they were just queried!`
        );
        return;
      }
      console.vlog("Loading apps");
      commit("setAppsQTime", new Date());
      let res = await api("GET", "api/apps", {
        icon: true,
      });
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description,
        });
        if (q !== undefined && q.hasOwnProperty("callback")) {
          q.callback();
        }
        return;
      }
      let cmap = {};
      res.data.map((v) => {
        cmap[v.id] = v;
      });
      commit("setApps", cmap);
      if (q !== undefined && q.hasOwnProperty("callback")) {
        q.callback();
      }
    },
    getUpdates: async function ({ commit }) {
      console.vlog("Checking for updates");
      let res = await api("GET", "api/server/updates");
      if (!res.response.ok) {
      } else {
        commit("setUpdates", res.data);
      }
    },
    getPluginApps: async function ({ commit, state }) {
      if (state.plugin_apps !== null) {
        return;
      }
      let res = await api("GET", "api/server/apps");
      if (!res.response.ok) {
      } else {
        commit("setPluginApps", res.data);
      }
    },
    ReadUserSettingsSchema: async function ({ commit, state, rootState }) {
      if (state.user_settings_schema != null) {
        return; //Already have it, no need to query again.
      }
      let res = await api("GET", `api/users/${encodeURIComponent(rootState.app.info.user.username)}/settings_schema`);
      if (!res.response.ok) {
      } else {
        commit("setUserSettingsSchema", res.data);
      }
    }
  },
};
