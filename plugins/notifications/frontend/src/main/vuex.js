import Vue from "../../dist.mjs";
import api from "../../api.mjs";

export default {
    state: {
        global: null
    },
    mutations: {
        setGlobalNotifications(state, v) {
            state.global = v;
        }
    },
    actions: {
        readGlobalNotifications: async function ({
            commit
        }) {
            let res = await api("GET", `api/heedy/v1/notifications`, {
                global: true
            });
            if (!res.response.ok) {
                commit("alert", {
                    type: "error",
                    text: res.data.error_description
                });

            } else {
                commit("setGlobalNotifications", res.data);
            }
        }
    }
};