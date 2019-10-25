var sourceRoutesMap = {};
var sourceRoutes = [];

class Source {
    constructor(app) {
        this.store = app.store;

        let querySource = (e) => {
            if (this.store.state.heedy.sources[e.source] !== undefined || this.store.state.heedy.userSources[e.user] !== undefined || e.connection !== undefined && this.store.state.heedy.connectionSources[e.connection] !== undefined) {
                this.store.dispatch("readSource_", {
                    id: e.source
                });
            }
        }
        // Subscribe to all source events, so that the source list
        // can be kept up-to-date
        if (app.info.user != null) {
            app.websocket.subscribe("source_create", {
                event: "source_create",
                user: app.info.user.username
            }, querySource);
            app.websocket.subscribe("source_update", {
                event: "source_update",
                user: app.info.user.username
            }, querySource);
            app.websocket.subscribe("source_delete", {
                event: "source_delete",
                user: app.info.user.username
            }, (e) => {
                if (this.store.state.heedy.sources[e.source] !== undefined || this.store.state.heedy.userSources[e.user] !== undefined || e.connection !== undefined && this.store.state.heedy.connectionSources[e.connection] !== undefined) {
                    this.store.commit("setSource", {
                        id: e.source,
                        isNull: true
                    });
                }
            });
        }

    }

    /**
     * Identical to a menu item, it is displayed in a special source creation menu
     * @param {*} c The creator to add
     */
    addCreator(c) {
        this.store.commit("addSourceCreator", c);
    }
    addComponent(c) {
        this.store.commit("addSourceComponent", c);
    }
    addType(c) {
        this.store.commit("addSourceType", c);
    }
    /**
     * Replace the page shown for the given source type with a custom component
     * @param {*} t The source type
     * @param {*} c Custom component to use
     */
    replacePage(t, c) {
        this.store.commit("addSourceCustomPage", {
            t: t,
            c: c
        });
    }

    /**
     * Adds a route to sources. It
     * automatically takes /source/:sourceid/{r.path}.
     * If the route works only on a specific source type, it is recommended to
     * prefix it with the type, ie: p.path = /stream/...
     * @param {*} r 
     */
    addRoute(r) {
        sourceRoutesMap[r.path] = r;
    }

    $onInit() {
        // Need to set the sourceRoutes with the right values:
        Object.values(sourceRoutesMap).reduce((_, r) => {
            if (r.path.startsWith("/")) {
                r.path = r.path.substring(1, r.path.length);
            }
            sourceRoutes.push(r);
            return null;
        }, null);
    }
}

export {
    sourceRoutes
}
export default Source;