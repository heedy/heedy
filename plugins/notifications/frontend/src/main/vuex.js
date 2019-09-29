import Vue from "../../dist.mjs";
import api from "../../api.mjs";

export default {
    state: {
        global: null,
        connections: {},
        sources: {}
    },
    mutations: {
        setGlobalNotifications(state, v) {
            v.sort((a, b) => b.timestamp - a.timestamp);
            state.global = v;
        },
        setConnectionNotifications(state, v) {
            let nmap = v.data.reduce((map, o) => {
                map[o.key] = o;
                return map;
            }, {});
            Vue.set(state.connections, v.id, nmap);
        },
        setSourceNotifications(state, v) {
            let nmap = v.data.reduce((map, o) => {
                map[o.key] = o;
                return map;
            }, {});
            Vue.set(state.sources, v.id, nmap);
        }
    },
    actions: {
        readGlobalNotifications: async function ({
            commit
        }) {
            if (this.debounce) {
                return;
            }
            this.debounce = true;
            console.log("Reading global notifications");
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
            this.debounce = false;
        },
        readConnectionNotifications: async function ({
            commit
        }, q) {
            if (this.debounce) {
                return;
            }
            this.debounce = true;
            console.log("Reading notifications for", q.id);
            let res = await api("GET", `api/heedy/v1/notifications`, {
                connection: q.id
            });
            if (!res.response.ok) {
                commit("alert", {
                    type: "error",
                    text: res.data.error_description
                });

            } else {
                commit("setConnectionNotifications", {
                    id: q.id,
                    data: res.data
                });
            }
            this.debounce = false;
        },
        readSourceNotifications: async function ({
            commit
        }, q) {
            if (this.debounce) {
                return;
            }
            this.debounce = true;
            console.log("Reading notifications for", q.id);
            let res = await api("GET", `api/heedy/v1/notifications`, {
                source: q.id
            });
            if (!res.response.ok) {
                commit("alert", {
                    type: "error",
                    text: res.data.error_description
                });

            } else {
                commit("setSourceNotifications", {
                    id: q.id,
                    data: res.data
                });
            }
            this.debounce = false;
        },
        updateNotification: async function ({
            commit
        }, q) {
            console.log("Updating notification", q);
            let res = await api("PATCH", `api/heedy/v1/notifications`, q.u, true, q.n);
            if (!res.response.ok) {
                commit("alert", {
                    type: "error",
                    text: res.data.error_description
                });

            }
        },
        deleteNotification: async function ({
            commit
        }, q) {
            console.log("DELETING notification", q);
            let res = await api("DELETE", `api/heedy/v1/notifications`, q);
            if (!res.response.ok) {
                commit("alert", {
                    type: "error",
                    text: res.data.error_description
                });

            }
        },
    }
};