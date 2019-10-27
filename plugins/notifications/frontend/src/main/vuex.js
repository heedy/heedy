import Vue from "../../dist/vue.mjs";
import moment from "../../dist/moment.mjs";
import api from "../../api.mjs";

// The notification key
function nKey(n) {
    return `${n.key}.${n.user}.${n.connection}.${n.source}`
}

export default {
    state: {
        global: null,
        global_qtime: null,
        connections: {},
        connections_qtime: {},
        sources: {},
        sources_qtime: {}
    },
    mutations: {
        setNotification(state, n) {
            if (state.global[nKey(n)] !== undefined || n.global) {
                if (!n.global) {
                    Vue.delete(state.global, nKey(n))
                } else {
                    Vue.set(state.global, nKey(n), n);
                }

            }
            if (n.source !== undefined) {
                if (state.sources[n.source] !== undefined) {
                    Vue.set(state.sources[n.source], n.key, n);
                }
                return
            }
            if (n.connection !== undefined) {
                if (state.connections[n.connection] !== undefined) {
                    Vue.set(state.connections[n.connection], n.key, n);
                }

                return;
            }
        },
        deleteNotification(state, n) {
            if (state.global[nKey(n)] !== undefined) {
                Vue.delete(state.global, nKey(n));
            }

            if (n.source !== undefined) {
                if (state.sources[n.source] !== undefined && state.sources[n.source][n.key] !== undefined) {
                    Vue.delete(state.sources[n.source], n.key);
                }
                return
            }
            if (n.connection !== undefined) {
                if (state.connections[n.connection] !== undefined && state.connections[n.connection][n.key] !== undefined) {
                    Vue.delete(state.connections[n.connection], n.key);
                }
                return;
            }
        },

        setGlobalNotifications(state, v) {
            let qtime = moment();
            // Turn a list of notifications into an object keyed by nKey
            state.global = v.reduce((o, n) => {
                n.qtime = qtime;
                o[nKey(n)] = n;
                return o;
            }, {});
            state.global_qtime = qtime;

            // Make sure to update all relevant notifications in the sources and connections
            v.forEach((n) => {
                if (n.source !== undefined) {
                    if (state.sources[n.source] !== undefined) {
                        Vue.set(state.sources, n.key, n);
                    }

                    return;
                }
                if (n.connection !== undefined) {
                    if (state.connections[n.connection] !== undefined) {
                        Vue.set(state.connections, n.key, n);
                    }

                    return;
                }
            });

        },
        setConnectionNotifications(state, v) {
            let qtime = moment();
            let nmap = v.data.reduce((map, o) => {
                o.qtime = qtime;
                map[o.key] = o;
                return map;
            }, {});
            Vue.set(state.connections, v.id, nmap);
            Vue.set(state.connections_qtime, v.id, qtime);
        },
        setSourceNotifications(state, v) {
            let qtime = moment();
            let nmap = v.data.reduce((map, o) => {
                o.qtime = qtime;
                map[o.key] = o;
                return map;
            }, {});
            Vue.set(state.sources, v.id, nmap);
            Vue.set(state.sources_qtime, v.id, qtime);
        }
    },
    actions: {
        readGlobalNotifications: async function ({
            commit,
            state,
            rootState
        }) {
            if (state.global != null && rootState.app.websocket != null && rootState.app.websocket.isBefore(state.global_qtime)) {
                console.log("Not querying global notifications - websocket active");
                return;
            }
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
        },
        readConnectionNotifications: async function ({
            commit,
            state,
            rootState
        }, q) {
            if (state.connections[q.id] !== undefined && rootState.app.websocket != null && rootState.app.websocket.isBefore(state.connections_qtime[q.id])) {
                console.log(`Not querying notifications for ${q.id} - websocket active`);
                return;
            }
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
        },
        readSourceNotifications: async function ({
            commit,
            state,
            rootState
        }, q) {
            if (state.sources[q.id] !== undefined && rootState.app.websocket != null && rootState.app.websocket.isBefore(state.sources_qtime[q.id])) {
                console.log(`Not querying notifications for ${q.id} - websocket active`);
                return;
            }
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