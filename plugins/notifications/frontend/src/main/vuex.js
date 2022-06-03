import Vue from "../../dist/vue.mjs";
import moment from "../../dist/moment.mjs";
import api from "../../util.mjs";

// The notification key
function nKey(n) {
    return `${n.key}.${n.user}.${n.app}.${n.object}`
}

export default {
    state: {
        global: null,
        global_qtime: null,
        apps: {},
        apps_qtime: {},
        objects: {},
        objects_qtime: {}
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
            if (n.object !== undefined) {
                if (state.objects[n.object] !== undefined) {
                    Vue.set(state.objects[n.object], n.key, n);
                }
                return
            }
            if (n.app !== undefined) {
                if (state.apps[n.app] !== undefined) {
                    Vue.set(state.apps[n.app], n.key, n);
                }

                return;
            }
        },
        deleteNotification(state, n) {
            if (state.global[nKey(n)] !== undefined) {
                Vue.delete(state.global, nKey(n));
            }

            if (n.object !== undefined) {
                if (state.objects[n.object] !== undefined && state.objects[n.object][n.key] !== undefined) {
                    Vue.delete(state.objects[n.object], n.key);
                }
                return
            }
            if (n.app !== undefined) {
                if (state.apps[n.app] !== undefined && state.apps[n.app][n.key] !== undefined) {
                    Vue.delete(state.apps[n.app], n.key);
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

            // Make sure to update all relevant notifications in the objects and apps
            v.forEach((n) => {
                if (n.object !== undefined) {
                    if (state.objects[n.object] !== undefined) {
                        Vue.set(state.objects, n.key, n);
                    }

                    return;
                }
                if (n.app !== undefined) {
                    if (state.apps[n.app] !== undefined) {
                        Vue.set(state.apps, n.key, n);
                    }

                    return;
                }
            });

        },
        setAppNotifications(state, v) {
            let qtime = moment();
            let nmap = v.data.reduce((map, o) => {
                o.qtime = qtime;
                map[o.key] = o;
                return map;
            }, {});
            Vue.set(state.apps, v.id, nmap);
            Vue.set(state.apps_qtime, v.id, qtime);
        },
        setObjectNotifications(state, v) {
            let qtime = moment();
            let nmap = v.data.reduce((map, o) => {
                o.qtime = qtime;
                map[o.key] = o;
                return map;
            }, {});
            Vue.set(state.objects, v.id, nmap);
            Vue.set(state.objects_qtime, v.id, qtime);
        }
    },
    actions: {
        readGlobalNotifications: async function ({
            commit,
            state,
            rootState
        }) {
            if (state.global != null && rootState.app.websocket != null && rootState.app.websocket.isBefore(state.global_qtime)) {
                console.vlog("Not querying global notifications - websocket active");
                return;
            }
            console.vlog("Reading global notifications");
            let res = await api("GET", `api/notifications`, {
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
        readAppNotifications: async function ({
            commit,
            state,
            rootState
        }, q) {
            if (state.apps[q.id] !== undefined && rootState.app.websocket != null && rootState.app.websocket.isBefore(state.apps_qtime[q.id])) {
                console.vlog(`Not querying notifications for ${q.id} - websocket active`);
                return;
            }
            console.vlog("Reading notifications for", q.id);
            let res = await api("GET", `api/notifications`, {
                app: q.id
            });
            if (!res.response.ok) {
                commit("alert", {
                    type: "error",
                    text: res.data.error_description
                });

            } else {
                commit("setAppNotifications", {
                    id: q.id,
                    data: res.data
                });
            }
        },
        readObjectNotifications: async function ({
            commit,
            state,
            rootState
        }, q) {
            if (state.objects[q.id] !== undefined && rootState.app.websocket != null && rootState.app.websocket.isBefore(state.objects_qtime[q.id])) {
                console.vlog(`Not querying notifications for ${q.id} - websocket active`);
                return;
            }
            console.vlog("Reading notifications for", q.id);
            let res = await api("GET", `api/notifications`, {
                object: q.id
            });
            if (!res.response.ok) {
                commit("alert", {
                    type: "error",
                    text: res.data.error_description
                });

            } else {
                commit("setObjectNotifications", {
                    id: q.id,
                    data: res.data
                });
            }
        },
        updateNotification: async function ({
            commit
        }, q) {
            console.vlog("Updating notification", q);
            let res = await api("PATCH", `api/notifications`, q.u, {params:q.n});
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
            console.vlog("DELETING notification", q);
            let res = await api("DELETE", `api/notifications`, null, {params:q});
            if (!res.response.ok) {
                commit("alert", {
                    type: "error",
                    text: res.data.error_description
                });

            }
        },
    }
};