class StreamInjector {
    constructor(app) {
        this.app = app;


        app.worker.addHandler("stream_query", (c, m) => this._onQuery(c, m))
    }

    addView(name, obj) {
        this.app.store.commit("addView", {
            key: name,
            component: obj
        });
    }

    async _onQuery(ctx, msg) {
        console.log("ReturnQ", ctx, msg);

        this.app.store.commit("setData", msg);
    }


}

export default StreamInjector;