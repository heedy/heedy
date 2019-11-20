class StreamInjector {
    constructor(app) {
        this.app = app;

        this.subscriptions = {};

        app.worker.addHandler("stream_views", (c, m) => this._onViews(c, m));

        // Watch the object objects, so that the worker always has the most recent
        // value. A more detailed explanation is in the worker.
        this.watchers = {};

    }

    addView(name, obj) {
        this.app.store.commit("addView", {
            key: name,
            component: obj
        });
    }

    _onViews(ctx, msg) {
        let skey = msg.id + ":" + msg.key;
        if (this.subscriptions[skey] === undefined) {
            console.error("Unknown stream view subscription key ", skey);
            return;
        }
        this.subscriptions[skey](msg.views);
    }

    subscribeQuery(stream, key, query, callback) {
        let skey = stream.id + ":" + key;
        this.subscriptions[skey] = callback;
        this.app.worker.postMessage("stream_subscribe_query", {
            stream: stream,
            key,
            query
        });
        if (this.watchers[stream.id] === undefined) {
            this.watchers[stream.id] = this.app.store.watch(
                (state, getters) => state.heedy.objects[stream.id],
                (n, o) => {
                    if (n === undefined || n === null) {
                        console.log("Stopping watch of ", stream.id);
                        this.watchers[stream.id]();
                        return;
                    }
                    this.app.worker.postMessage("stream_update", n)
                }
            )
        }
    }
    unsubscribeQuery(streamid, key) {
        this.app.worker.postMessage("stream_unsubscribe_query", {
            id: streamid,
            key
        });
        let skey = streamid + ":" + key;
        delete this.subscriptions[skey];
    }

    query(stream, query, callback) {
        let skey = stream.id + ":" + key;
        this.subscriptions[skey] = (d) => {
            delete this.subscriptions[skey];
            callback(d);
        };
        this.app.worker.postMessage("stream_query", {
            stream: stream,
            key,
            query
        });
    }

}

export default StreamInjector;