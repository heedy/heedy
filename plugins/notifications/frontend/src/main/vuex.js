import Vue from "../../dist.mjs";
import api from "../../api.mjs";

// The notification key
function nKey(n) {
    return `${n.key}.${n.user}.${n.connection}.${n.source}`
}

export default {
    state: {
        global: null,
        connections: {},
        sources: {}
    },
    mutations: {
        setGlobalNotifications(state, v) {
            state.global = v.reduce((o, n) => {
                o[nKey(n)] = n;
                return o;
            }, {});
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