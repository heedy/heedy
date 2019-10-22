class StreamInjector {
    constructor(wkr) {
        this.worker = wkr;
        this.processors = {};

        wkr.addHandler("stream_query", (ctx, msg) => this._onQuery(ctx, msg))
    }

    addProcessor(key, f) {
        this.processors[key] = f;
    }

    async _onQuery(ctx, msg) {
        console.log("Query", msg);



        let vals = Object.values(this.processors).map(v => v(msg.source, msg.data));
        let outvals = {};
        for (let i = 0; i < vals.length; i++) {
            let res = await vals[i];
            outvals = Object.assign(outvals, res);
        }
        console.log("THERE ARE", outvals);

        this.worker.postMessage("stream_query", {
            data: outvals,
            id: msg.source.id
        });
    }

}

export default StreamInjector;