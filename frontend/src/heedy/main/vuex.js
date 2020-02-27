import Vue from "../../dist/vue.mjs";
import moment from "../../dist/moment.mjs";
import api from "../../api.mjs";

export default {
  state: {
    alert: {
      value: false,
      text: "",
      type: ""
    },
    users: {},
    // Components to show for a user
    user_components: [],

    // The current user's apps & time when they were queried
    apps: null,
    apps_qtime: null,

    // The map of available plugin apps
    plugin_apps: null,

    // Components to show in the app
    app_components: [],

    // A map of objects
    objects: {},
    // Components to show for a object
    object_components: [],
    // The custom pages to show for the given object type
    object_custom_pages: {},
    // Object types tell how to list the objects
    object_types: {},

    // a map keyed by username, where each element is a map of ids to null
    userObjects: {},
    userObjects_qtime: {},

    // a map keyed by app id, where each element is a map of ids to null
    appObjects: {},
    appObjects_qtime: {},

    // The following are initialized by the objectInjector
    objectCreators: [],

    // Subpaths for each object type
    typePaths: {},

    // The map of app scopes along with their descriptions
    appScope: null,

    settings_routes: [],
    updates: {
      heedy: false,
      plugins: [],
      config: false
    }
  },
  mutations: {
    setSettingsRoutes(state, v) {
      state.settings_routes = v;
    },
    addAppComponent(state, v) {
      state.app_components.push(v);
    },
    addUserComponent(state, v) {
      state.user_components.push(v);
    },
    addObjectType(state, v) {
      state.object_types[v.type] = v;
    },
    addObjectComponent(state, v) {
      state.object_components.push(v);
    },
    addObjectCustomPage(state, p) {
      state.object_custom_pages[p.t] = p.c;
    },
    addObjectCreator(state, c) {
      state.objectCreators.push(c);
    },
    alert(state, v) {
      state.alert = {
        value: true,
        type: "",
        text: "",
        ...v
      };
    },
    setUser(state, v) {
      if (v.isNull !== undefined) {
        if (state.userObjects[v.username] !== undefined) {
          Vue.delete(state.userObjects, v.username);
        }
        Vue.set(state.users, v.username, null);
        return;
      }
      Vue.set(state.users, v.username, {
        qtime: moment(),
        ...v
      });
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
        qtime: moment(),
        ...v
      });
    },
    setApps(state, v) {
      let qtime = moment();
      Object.keys(v).forEach(k => {
        v[k] = {
          qtime,
          ...v[k]
        };
      });
      Object.keys(state.appObjects).forEach(k => {
        if (v[k] === undefined) {
          Vue.delete(state.appObjects, k);
        }
      });
      state.apps = v;
      state.apps_qtime = moment();
    },
    setObject(state, v) {
      // First check if the object has existing value
      let curs = state.objects[v.id] || null;
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
        return;
      }
      // Set the object
      Vue.set(state.objects, v.id, {
        qtime: moment(),
        ...v
      });
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
      if (state.userObjects[v.owner] !== undefined && v.app == null) {
        Vue.set(state.userObjects[v.owner], v.id, null);
      }
    },
    setUserObjects(state, v) {
      let srcidmap = {};
      let qtime = moment();
      v.objects.forEach(s => {
        srcidmap[s.id] = null;
        Vue.set(state.objects, s.id, {
          qtime,
          ...s
        });
      });
      Vue.set(state.userObjects, v.user, srcidmap);
      Vue.set(state.userObjects_qtime, v.user, qtime);
    },
    setAppObjects(state, v) {
      let srcidmap = {};
      let qtime = moment();
      v.objects.forEach(s => {
        srcidmap[s.id] = null;
        Vue.set(state.objects, s.id, {
          qtime,
          ...s
        });
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
    }
  },
  actions: {
    errnotify({
      commit
    }, v) {
      // Notifies of an error
      if (v.hasOwnProperty("error")) {
        // Only notify if it is an actual error
        commit("alert", {
          type: "error",
          text: v.error_description
        });
      }
    },
    // This function performs a query on the user, ignoring websocket
    readUser_: async function ({
      commit,
      rootState
    }, q) {
      let username = q.username;
      console.log("Reading user", username);
      let res = await api("GET", `api/users/${username}`, {
        icon: true
      });
      console.log(res);
      if (!res.response.ok) {
        // If the error is 404, set the user to null
        if (res.response.status == 400 || res.response.status == 403) {
          // TODO: 404 should be returned
          commit("setUser", {
            username: username,
            isNull: true
          });
        } else {
          commit("alert", {
            type: "error",
            text: res.data.error_description
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
    readApp_: async function ({
      commit
    }, q) {
      console.log("Reading app", q.id);
      let res = await api("GET", `api/apps/${q.id}`, {
        icon: true
      });
      if (!res.response.ok) {
        if (res.response.status == 400 || res.response.status == 403) {
          // TODO: 404 should be returned
          commit("setApp", {
            id: q.id,
            isNull: true
          });
        } else {
          commit("alert", {
            type: "error",
            text: res.data.error_description
          });
        }
      } else {
        commit("setApp", res.data);
      }

      if (q.hasOwnProperty("callback")) {
        q.callback();
      }
    },
    readObject_: async function ({
      commit
    }, q) {
      console.log("Reading object", q.id);
      let res = await api("GET", `api/objects/${q.id}`, {
        icon: true
      });
      if (!res.response.ok) {
        if (res.response.status == 400 || res.response.status == 403) {
          // TODO: 404 should be returned
          commit("setObject", {
            id: q.id,
            isNull: true
          });
        } else {
          commit("alert", {
            type: "error",
            text: res.data.error_description
          });
        }
      } else {
        commit("setObject", res.data);
      }

      if (q.hasOwnProperty("callback")) {
        q.callback();
      }
    },

    readUser({
      state,
      rootState,
      dispatch
    }, q) {
      let username = q.username;
      if (
        state.users[username] !== undefined &&
        state.users[username] != null
      ) {
        if (
          rootState.app.websocket != null &&
          rootState.app.websocket.isBefore(state.users[username].qtime)
        ) {
          console.log(`Not querying ${username} - websocket active`);
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
          return;
        }
      }
      dispatch("readUser_", q);
    },
    readApp: async function ({
      state,
      rootState,
      dispatch
    }, q) {
      if (state.apps == null) {
        dispatch("listApps", q);
        return;
      }
      if (
        state.apps[q.id] !== undefined &&
        rootState.app.websocket != null &&
        rootState.app.websocket.isBefore(state.apps[q.id].qtime)
      ) {
        console.log(`Not querying ${q.id} - websocket active`);
        if (q.hasOwnProperty("callback")) {
          q.callback();
        }
        return;
      }
      dispatch("readApp_", q);
    },
    readObject: async function ({
      state,
      rootState,
      dispatch
    }, q) {
      if (state.objects[q.id] !== undefined && state.objects[q.id] !== null) {
        if (
          rootState.app.websocket != null &&
          rootState.app.websocket.isBefore(state.objects[q.id].qtime)
        ) {
          console.log(`Not querying ${q.id} - websocket active`);
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
          return;
        }
      }
      dispatch("readObject_", q);
    },
    readUserObjects: async function ({
      commit,
      state,
      rootState
    }, q) {
      // Only if they are not being kept up-to-date by the websocket
      if (
        state.userObjects[q.username] !== undefined &&
        rootState.app.websocket !== null &&
        rootState.app.websocket.isBefore(state.userObjects_qtime[q.username])
      ) {
        console.log(`Not reading ${q.username} objects - websocket active`);
        return;
      }
      console.log("Reading objects for user", q.username);
      let query = {
        owner: q.username,
        icon: true
      };

      if (
        rootState.app.info.user != null &&
        rootState.app.info.user.username == q.username
      ) {
        query["app"] = "";
      }

      let res = await api("GET", `api/objects`, query);
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description
        });
      } else {
        commit("setUserObjects", {
          user: q.username,
          objects: res.data
        });
      }

      if (q.hasOwnProperty("callback")) {
        q.callback();
      }
    },
    readAppObjects: async function ({
      commit,
      state,
      rootState
    }, q) {
      // Only if they are not being kept up-to-date by the websocket
      if (
        state.appObjects[q.id] !== undefined &&
        rootState.app.websocket !== null &&
        rootState.app.websocket.isBefore(state.appObjects_qtime[q.id])
      ) {
        console.log(`Not reading ${q.id} objects - websocket active`);
        return;
      }
      console.log("Reading objects for app", q.id);
      let query = {
        app: q.id,
        icon: true
      };

      let res = await api("GET", `api/objects`, query);
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description
        });
      } else {
        commit("setAppObjects", {
          id: q.id,
          objects: res.data
        });
      }

      if (q.hasOwnProperty("callback")) {
        q.callback();
      }
    },
    getAppScope: async function ({
      commit
    }) {
      console.log("Loading available app scopes");
      let res = await api("GET", "api/server/scope");
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description
        });
      } else {
        commit("setAppScope", res.data);
      }
    },
    listApps: async function ({
      commit,
      state,
      rootState
    }, q) {
      // Only list apps if they are not being kept up-to-date by the websocket
      if (
        state.apps !== null &&
        rootState.app.websocket !== null &&
        rootState.app.websocket.isBefore(state.apps_qtime)
      ) {
        console.log("Not listing apps - websocket active");
        return;
      }
      console.log("Loading apps");
      let res = await api("GET", "api/apps", {
        icon: true
      });
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description
        });
        if (q !== undefined && q.hasOwnProperty("callback")) {
          q.callback();
        }
        return;
      }
      let cmap = {};
      res.data.map(v => {
        cmap[v.id] = v;
      });
      commit("setApps", cmap);
      if (q !== undefined && q.hasOwnProperty("callback")) {
        q.callback();
      }
    },

    getUpdates: async function ({
      commit
    }) {
      console.log("Checking if updates ready");
      let res = await api("GET", "api/server/updates");
      if (!res.response.ok) {} else {
        commit("setUpdates", res.data);
      }
    },
    getPluginApps: async function ({
      commit,
      state
    }) {
      if (state.plugin_apps !== null) {
        return;
      }
      let res = await api("GET", "api/server/apps");
      if (!res.response.ok) {} else {
        commit("setPluginApps", res.data);
      }
    }
  }
};