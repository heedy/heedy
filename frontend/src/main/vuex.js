import Vue, {
    createLogger
} from "../dist/vue.mjs";

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
                }
            }
        },
        plugins: ((_DEBUG || appinfo.verbose) ? [createLogger()] : [])
    }
};


export default setup;