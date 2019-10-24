import Vue from "../../dist/vue.mjs";
import moment from "../../dist/moment.mjs";
import api from "../../api.mjs";

export default {
  state: {
    // The status of the websocket. null means disconnected, and a moment() object
    // gives the time from which it was connected
    websocket: null,

    alert: {
      value: false,
      text: "",
      type: ""
    },
    users: {},
    // Components to show for a user
    user_components: [],

    // The current user's connections & time when they were queried
    connections: null,
    connections_qtime: null,

    // Components to show in the connection
    connection_components: [],

    // A map of sources
    sources: {},
    // Components to show for a source
    source_components: [],
    // The custom pages to show for the given source type
    source_custom_pages: {},

    // a map keyed by username, where each element is a map of ids to null
    userSources: {},
    userSources_qtime: {},

    // a map keyed by connection id, where each element is a map of ids to null
    connectionSources: {},
    connectionSources_qtime: {},

    // The following are initialized by the sourceInjector
    sourceCreators: [],

    // Subpaths for each source type
    typePaths: {},

    // The map of connection scopes along with their descriptions
    connectionScopes: null,

    settings_routes: [],
    updates: {
      heedy: false,
      plugins: [],
      config: false
    }
  },
  mutations: {
    setWebsocket(state, v) {
      state.websocket = v;
    },
    setSettingsRoutes(state, v) {
      state.settings_routes = v;
    },
    addConnectionComponent(state, v) {
      state.connection_components.push(v);
    },
    addUserComponent(state, v) {
      state.user_components.push(v);
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
    setConnection(state, v) {
      if (state.connections == null) {
        state.connections = {};
      }
      if (v.isNull !== undefined) {
        if (state.connectionSources[v.id] !== undefined) {
          Vue.delete(state.connectionSources, v.id);
        }
        if (state.connections[v.id] !== undefined) {
          Vue.delete(state.connections, v.id);
        }
        return
      }
      Vue.set(state.connections, v.id, {
        qtime: moment(),
        ...v
      });
    },
    setConnections(state, v) {
      let qtime = moment();
      Object.keys(v).forEach(k => {
        v[k] = {
          qtime,
          ...v[k]
        };
      });
      Object.keys(state.connectionSources).forEach(k => {
        if (v[k] === undefined) {
          Vue.delete(state.connectionSources, k);
        }
      })
      state.connections = v;
      state.connections_qtime = moment();
    },
    setSource(state, v) {
      // First check if the source has existing value
      let curs = state.sources[v.id] || null;
      if (v.isNull !== undefined) {
        // The source is to be deleted - make sure to take care of all places it could be
        if (curs !== null) {
          if (curs.connection !== null) {
            if (state.connectionSources[curs.connection] !== undefined) {
              Vue.delete(state.connectionSources[curs.connection], curs.id);
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
        if (v.connection != curs.connection) {
          if (state.connectionSources[curs.connection] !== undefined) {
            Vue.delete(state.connectionSources[curs.connection], curs.id);
          }
        }
        if (v.owner != curs.owner) {
          if (state.userSources[curs.owner] !== undefined) {
            Vue.delete(state.userSources[curs.owner], curs.id);
          }
        }
      }
      // Make sure to set it in the appropriate lists
      if (v.connection != null && state.connectionSources[v.connection] !== undefined) {
        Vue.set(state.connectionSources[v.connection], v.id, null);
      }
      if (state.userSources[v.owner] !== undefined) {
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
    setConnectionSources(state, v) {
      let srcidmap = {};
      let qtime = moment();
      v.sources.forEach(s => {
        srcidmap[s.id] = null;
        Vue.set(state.sources, s.id, {
          qtime,
          ...s
        });
      });
      Vue.set(state.connectionSources, v.id, srcidmap);
      Vue.set(state.connectionSources_qtime, v.id, qtime);
    },
    setConnectionScopes(state, v) {
      state.connectionScopes = v;
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
    readConnection_: async function ({
      commit
    }, q) {
      console.log("Reading connection", q.id);
      let res = await api("GET", `api/heedy/v1/connections/${q.id}`, {
        icon: true
      });
      if (!res.response.ok) {
        if (res.response.status == 400 || res.response.status == 403) { // TODO: 404 should be returned
          commit("setConnection", {
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
        commit("setConnection", res.data);
      }


      if (q.hasOwnProperty("callback")) {
        q.callback();
      }
    },
    readSource_: async function ({
      commit,
      rootState
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
      dispatch
    }, q) {
      let username = q.username;
      if (state.users[username] !== undefined && state.users[username] != null) {
        if (state.websocket != null && state.websocket.isBefore(state.users[username].qtime)) {
          console.log(`Not querying ${username} - websocket active`);
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
          return;
        }
      }
      dispatch("readUser_", q);


    },
    readConnection: async function ({
      state,
      dispatch
    }, q) {
      if (state.connections == null) {
        dispatch("listConnections", q);
        return;
      }
      if (state.connections[q.id] !== undefined && state.websocket != null && state.websocket.isBefore(state.connections[q.id].qtime)) {
        console.log(`Not querying ${q.id} - websocket active`);
        if (q.hasOwnProperty("callback")) {
          q.callback();
        }
        return;
      }
      dispatch("readConnection_", q);
    },
    readSource: async function ({
      state,
      dispatch
    }, q) {
      if (state.sources[q.id] !== undefined && state.sources[q.id] !== null) {
        if (state.websocket != null && state.websocket.isBefore(state.sources[q.id].qtime)) {
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
      if (state.userSources[q.username] !== undefined && state.websocket !== null && state.websocket.isBefore(state.userSources_qtime[q.username])) {
        console.log(`Not reading ${q.username} sources - websocket active`);
        return;
      }
      console.log("Reading sources for user", q.username);
      let query = {
        username: q.username
      };

      if (rootState.app.info.user != null && rootState.app.info.user.username == q.username) {
        query["connection"] = "none";
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
    readConnectionSources: async function ({
      commit,
      state
    }, q) {
      // Only if they are not being kept up-to-date by the websocket
      if (state.connectionSources[q.id] !== undefined && state.websocket !== null && state.websocket.isBefore(state.connectionSources_qtime[q.id])) {
        console.log(`Not reading ${q.id} sources - websocket active`);
        return;
      }
      console.log("Reading sources for connection", q.id);
      let query = {
        connection: q.id
      };


      let res = await api("GET", `api/heedy/v1/sources`, query);
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description
        });

      } else {
        commit("setConnectionSources", {
          id: q.id,
          sources: res.data
        });
      }


      if (q.hasOwnProperty("callback")) {
        q.callback();
      }

    },
    getConnectionScopes: async function ({
      commit
    }) {
      console.log("Loading available connection scopes");
      let res = await api("GET", "api/heedy/v1/server/scopes");
      if (!res.response.ok) {
        commit("alert", {
          type: "error",
          text: res.data.error_description
        });

      } else {
        commit("setConnectionScopes", res.data);
      }
    },
    listConnections: async function ({
      commit,
      state
    }, q) {
      // Only list connections if they are not being kept up-to-date by the websocket
      if (state.connections !== null && state.websocket !== null && state.websocket.isBefore(state.connections_qtime)) {
        console.log("Not listing connections - websocket active");
        return;
      }
      console.log("Loading connections");
      let res = await api("GET", "api/heedy/v1/connections", {
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
      commit("setConnections", cmap);
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