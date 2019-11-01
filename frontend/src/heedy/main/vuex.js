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

    // Components to show in the app
    app_components: [],

    // A map of sources
    sources: {},
    // Components to show for a source
    source_components: [],
    // The custom pages to show for the given source type
    source_custom_pages: {},
    // Source types tell how to list the sources
    source_types: {},

    // a map keyed by username, where each element is a map of ids to null
    userSources: {},
    userSources_qtime: {},

    // a map keyed by app id, where each element is a map of ids to null
    appSources: {},
    appSources_qtime: {},

    // The following are initialized by the sourceInjector
    sourceCreators: [],

    // Subpaths for each source type
    typePaths: {},

    // The map of app scopes along with their descriptions
    appScopes: null,

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
    addSourceType(state, v) {
      state.source_types[v.type] = v;
    },
    addSourceComponent(state, v) {
      state.source_components.push(v);
    },
    addSourceCustomPage(state, p) {
      state.source_custom_pages[p.t] = p.c;
    },
    addSourceCreator(state, c) {
      state.sourceCreators.push(c);
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
        if (state.userSources[v.username] !== undefined) {
          Vue.delete(state.userSources, v.username);
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
        if (state.appSources[v.id] !== undefined) {
          Vue.delete(state.appSources, v.id);
        }
        if (state.apps[v.id] !== undefined) {
          Vue.delete(state.apps, v.id);
        }
        return
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
      Object.keys(state.appSources).forEach(k => {
        if (v[k] === undefined) {
          Vue.delete(state.appSources, k);
        }
      })
      state.apps = v;
      state.apps_qtime = moment();
    },
    setSource(state, v) {
      // First check if the source has existing value
      let curs = state.sources[v.id] || null;
      if (v.isNull !== undefined) {
        // The source is to be deleted - make sure to take care of all places it could be
        if (curs !== null) {
          if (curs.app !== null) {
            if (state.appSources[curs.app] !== undefined) {
              Vue.delete(state.appSources[curs.app], curs.id);
            }
          } else if (state.userSources[curs.owner] !== undefined) {
            Vue.delete(state.userSources[curs.owner], curs.id);
          }
        }
        Vue.set(state.sources, v.id, null);
        return;
      }
      // Set the source
      Vue.set(state.sources, v.id, {
        qtime: moment(),
        ...v
      });
      // Delete from lists where changed
      if (curs != null) {
        if (v.app != curs.app) {
          if (state.appSources[curs.app] !== undefined) {
            Vue.delete(state.appSources[curs.app], curs.id);
          }
        }
        if (v.owner != curs.owner) {
          if (state.userSources[curs.owner] !== undefined) {
            Vue.delete(state.userSources[curs.owner], curs.id);
          }
        }
      }
      // Make sure to set it in the appropriate lists
      if (v.app != null && state.appSources[v.app] !== undefined) {
        Vue.set(state.appSources[v.app], v.id, null);
      }
      if (state.userSources[v.owner] !== undefined && v.app == null) {
        Vue.set(state.userSources[v.owner], v.id, null);
      }
    },
    setUserSources(state, v) {
      let srcidmap = {};
      let qtime = moment();
      v.sources.forEach(s => {
        srcidmap[s.id] = null;
        Vue.set(state.sources, s.id, {
          qtime,
          ...s
        });
      });
      Vue.set(state.userSources, v.user, srcidmap);
      Vue.set(state.userSources_qtime, v.user, qtime);
    },
    setAppSources(state, v) {
      let srcidmap = {};
      let qtime = moment();
      v.sources.forEach(s => {
        srcidmap[s.id] = null;
        Vue.set(state.sources, s.id, {
          qtime,
          ...s
        });
      });
      Vue.set(state.appSources, v.id, srcidmap);
      Vue.set(state.appSources_qtime, v.id, qtime);
    },
    setAppScopes(state, v) {
      state.appScopes = v;
    },
    setUpdates(state, v) {
      state.updates = v;
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
      let res = await api("GET", `api/heedy/v1/users/${username}`, {
        icon: true
      });
      console.log(res);
      if (!res.response.ok) {
        // If the error is 404, set the user to null
        if (res.response.status == 400 || res.response.status == 403) { // TODO: 404 should be returned
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
        if (rootState.app.info.user != null && rootState.app.info.user.username == username) {
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
      let res = await api("GET", `api/heedy/v1/apps/${q.id}`, {
        icon: true
      });
      if (!res.response.ok) {
        if (res.response.status == 400 || res.response.status == 403) { // TODO: 404 should be returned
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
    readSource_: async function ({
      commit
    }, q) {
      console.log("Reading source", q.id);
      let res = await api("GET", `api/heedy/v1/sources/${q.id}`, {
        icon: true
      });
      if (!res.response.ok) {
        if (res.response.status == 400 || res.response.status == 403) { // TODO: 404 should be returned
          commit("setSource", {
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
        commit("setSource", res.data);
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
      if (state.users[username] !== undefined && state.users[username] != null) {
        if (rootState.app.websocket != null && rootState.app.websocket.isBefore(state.users[username].qtime)) {
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
      if (state.apps[q.id] !== undefined && rootState.app.websocket != null && rootState.app.websocket.isBefore(state.apps[q.id].qtime)) {
        console.log(`Not querying ${q.id} - websocket active`);
        if (q.hasOwnProperty("callback")) {
          q.callback();
        }
        return;
      }
      dispatch("readApp_", q);
    },
    readSource: async function ({
      state,
      rootState,
      dispatch
    }, q) {
      if (state.sources[q.id] !== undefined && state.sources[q.id] !== null) {
        if (rootState.app.websocket != null && rootState.app.websocket.isBefore(state.sources[q.id].qtime)) {
          console.log(`Not querying ${q.id} - websocket active`);
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
          return;
        }
      }
      dispatch("readSource_", q);
    },
    readUserSources: async function ({
      commit,
      state,
      rootState
    }, q) {
      // Only if they are not being kept up-to-date by the websocket
      if (state.userSources[q.username] !== undefined && rootState.app.websocket !== null && rootState.app.websocket.isBefore(state.userSources_qtime[q.username])) {
        console.log(`Not reading ${q.username} sources - websocket active`);
        return;
      }
      console.log("Reading sources for user", q.username);
      let query = {
        username: q.username
      };

      if (rootState.app.info.user != null && rootState.app.info.user.username == q.username) {
        query["app"] = "none";
      }

      let res = await api("GET", `api/heedy/v1/sources`, query);
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description
        });

      } else {
        commit("setUserSources", {
          user: q.username,
          sources: res.data
        });
      }


      if (q.hasOwnProperty("callback")) {
        q.callback();
      }

    },
    readAppSources: async function ({
      commit,
      state,
      rootState
    }, q) {
      // Only if they are not being kept up-to-date by the websocket
      if (state.appSources[q.id] !== undefined && rootState.app.websocket !== null && rootState.app.websocket.isBefore(state.appSources_qtime[q.id])) {
        console.log(`Not reading ${q.id} sources - websocket active`);
        return;
      }
      console.log("Reading sources for app", q.id);
      let query = {
        app: q.id
      };


      let res = await api("GET", `api/heedy/v1/sources`, query);
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description
        });

      } else {
        commit("setAppSources", {
          id: q.id,
          sources: res.data
        });
      }


      if (q.hasOwnProperty("callback")) {
        q.callback();
      }

    },
    getAppScopes: async function ({
      commit
    }) {
      console.log("Loading available app scopes");
      let res = await api("GET", "api/heedy/v1/server/scopes");
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description
        });

      } else {
        commit("setAppScopes", res.data);
      }
    },
    listApps: async function ({
      commit,
      state,
      rootState
    }, q) {
      // Only list apps if they are not being kept up-to-date by the websocket
      if (state.apps !== null && rootState.app.websocket !== null && rootState.app.websocket.isBefore(state.apps_qtime)) {
        console.log("Not listing apps - websocket active");
        return;
      }
      console.log("Loading apps");
      let res = await api("GET", "api/heedy/v1/apps", {
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
        cmap[v.id] = v
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
      let res = await api("GET", "api/heedy/v1/server/updates");
      if (!res.response.ok) {} else {
        commit("setUpdates", res.data);
      }
    }

  }

};