

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
        users: users
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
          }
          if (rootState.app.info.user!=null && rootState.app.info.user.name == q.name) {
            commit("updateLoggedInUser",res.data);
          }
          commit("setUser", res.data);
          
          if (q.hasOwnProperty("callback")) {
            q.callback();
          }
        }
      }
};