
import Vue from "../../dist.mjs";
import api from "../api.mjs";

let users = {};
if (appinfo.user != null) {
  users[appinfo.user.name] = appinfo.user;
}

export default {
    state: {
        alert: {
            value: false,
            text: "",
            type: ""
        },
        users: users,
        sources: {},

        // A list of IDs under each user's key
        userSources: {},

        // The following are initialized by the sourceInjector
        sourceCreators: []
    },
    mutations: {
        alert(state, v) {
          state.alert = {
            value: true,
            type: "",
            text: "",
            ...v
          };
        },
        setUser(state, v) {
          Vue.set(state.users, v.name, v);
        },
        setSource(state, v) {
          Vue.set(state.sources,v.id, v);
        },
        setUserSources(state,v) {
          let srcidarray = [];
          for (let i=0;i < v.sources.length;i++) {
            Vue.set(state.sources,v.sources[i].id, v.sources[i]);
            srcidarray.push(v.sources[i].id);
          }
          Vue.set(state.userSources,v.user,srcidarray);
        }
      },
      actions: {
        errnotify({ commit }, v) {
          // Notifies of an error
          if (v.hasOwnProperty("error")) {
            // Only notify if it is an actual error
            commit("alert", {
              type: "error",
              text: v.error_description
            });
          }
        },
        readUser: async function({ commit, rootState }, q) {
          console.log("Reading user", q.name);
          let res = await api("GET", `api/heedy/v1/user/${q.name}`, {
            avatar: true
          });
          if (!res.response.ok) {
            commit("alert", {
              type: "error",
              text: res.data.error_description
            });
            
          } else {
            if (rootState.app.info.user!=null && rootState.app.info.user.name == q.name) {
              commit("updateLoggedInUser",res.data);
            }
            commit("setUser", res.data);
          }
          
          
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
      },
      readSource: async function({ commit, rootState }, q) {
        console.log("Reading source", q.id);
        let res = await api("GET", `api/heedy/v1/source/${q.id}`, {
          avatar: true
        });
        if (!res.response.ok) {
          commit("alert", {
            type: "error",
            text: res.data.error_description
          });
          
        } else {
          commit("setSource", res.data);
        }
        
        
        if (q.hasOwnProperty("callback")) {
          q.callback();
        }
      },
      readUserSources: async function({commit,rootState}, q) {
        console.log("Reading sources for user", q.name);
        let query = {user: q.name};

        if (rootState.app.info.user!=null && rootState.app.info.user.name == q.name) {
          query["connection"] = "none";
        }

        let res = await api("GET", `api/heedy/v1/source`, query);
        if (!res.response.ok) {
          commit("alert", {
            type: "error",
            text: res.data.error_description
          });
          
        } else {
          commit("setUserSources", {user: q.name, sources: res.data});
        }
        
        
        if (q.hasOwnProperty("callback")) {
          q.callback();
        }

      }

    }
};