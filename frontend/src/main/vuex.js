import Vue, {
    createLogger
} from "../dist/vue.mjs";

import api from "../util.mjs";

function setup(appinfo) {
    return {
        modules: {
            app: {
                state: {
                    info: appinfo,
                    // menu_items gives all the defined menu items
                    menu_items: {},
                    // The status of the websocket. null means disconnected, and a moment() object
                    // gives the time from which it was connected
                    websocket: null,
                    // Whether to notify the user that an update is available to the frontend
                    update_available: false,
                },
                mutations: {
                    updateLoggedInUser(state, v) {
                        state.info.user = v;
                    },
                    updateAppInfo(state,v) {
                        state.info = v;
                    },
                    UpdateUserPluginSettings(state, v) {
                        Vue.set(state.info.settings, v.plugin, v.value);
                    },
                    addMenuItem(state, m) {
                        state.menu_items[m.key] = m;
                    },
                    setWebsocket(state, v) {
                        state.websocket = v;
                    },
                    setUpdateAvailable(state, v) {
                        state.update_available = v;
                    },
                },
                actions: {
                    ReadUserPluginSettings: async function ({ commit, state }, q) {
                        let res = await api("GET", `api/users/${encodeURIComponent(state.info.user.username)}/settings/${encodeURIComponent(q.plugin)}`);
                        if (!res.response.ok) {
                        } else {
                            commit("UpdateUserPluginSettings", { plugin: q.plugin, value: res.data });
                        }
                    },
                    ReadAppInfo: async function({commit, state}) {
                        let res = await api("GET", "frontend_context");
                        if (!res.response.ok) {
                        } else {
                            // If either the version changed or plugins changed, or the user changed,
                            // ask for a page reload instead of replacing appinfo
                            let o1 = {version: res.data.version, plugins: res.data.plugins,admin:res.data.admin, user: state.info.user?.username};
                            let o2 = {version: state.info.version, plugins: state.info.plugins,admin:state.info.admin, user: state.info.user?.username};
                            if (JSON.stringify(o1) !== JSON.stringify(o2)) {
                                
                                commit("setUpdateAvailable", true);
                            } else {
                                commit("updateAppInfo", res.data);
                                commit("setUser", res.data.user); // Set the user in cache from appinfo
                            }

                            
                        }
                    }
                }
            }
        },
        plugins: ((_DEBUG || appinfo.verbose) ? [createLogger()] : [])
    }
};


export default setup;