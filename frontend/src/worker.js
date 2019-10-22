class Wrkr {
    constructor() {
        console.log("Running worker");
        this.handlers = {
            'import': (ctx, data) => this._importHandler(ctx, data)
        };
    }

    addHandler(key, f) {
        this.handlers[key] = f;
    }


    async _importHandler(ctx, msg) {
        try {
            (await import("./" + msg)).default(this);
        } catch (err) {
            console.error(err);
        }
    }
    postMessage(key, msg) {
        // post message
        postMessage({
            key: key,
            msg: msg
        });
    }

    async _onMessage(e) {
        let msg = e.data;
        console.log("Worker:", msg);
        if (this.handlers[msg.key] !== undefined) {
            let ctx = {
                key: msg.key,
            };
            await this.handlers[msg.key](ctx, msg.msg);
        } else {
            console.error(`Unknown message key ${msg.key}`);
        }
    }

    inject(name, p) {
        this[name] = p;
    }
}

let worker = new Wrkr();

// In the future this won't be necessary, since this will be a worker,
// but for now, we just emulate the worker.
export default worker;