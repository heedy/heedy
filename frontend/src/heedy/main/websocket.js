class EventSubscriber {
    constructor(retryConnect) {
        // Get subscriptions by key
        this.subscriptions = {};

        // Set up the websocket
        let wsproto = "wss:"
        if (location.protocol == "http:") {
            wsproto = "ws:"
        }

        this.loc = wsproto + "//" + location.host + location.pathname + "api/heedy/v1/events";

        this.retryConnect = retryConnect;
        this.isopen = false;

        this.resetTimeout = 200;
        this.retryTimeout = 200;
        this.retryTimeoutDelta = 1000;

        this.connect();
    }
    connect() {
        console.log(`Connecting to websocket ${this.loc}`);
        this.ws = new WebSocket(this.loc);

        this.ws.onopen = () => this.onopen();
        this.ws.onmessage = (m) => this.fire(m);
        this.ws.onclose = (e) => this.onclose(e);
    }
    onopen() {
        console.log("Websocket open");
        this.isopen = true;
        this.retryTimeout = this.resetTimeout;

        Object.values(this.subscriptions).forEach((s) => {
            let m = {
                cmd: "subscribe",
                ...s.event
            }
            console.log("<-", m);
            this.ws.send(JSON.stringify(m))
        });
    }
    onclose(e) {
        console.log("Websocket closed");
        this.isopen = false;
        if (this.retryConnect) {
            setTimeout(() => this.connect(), this.retryTimeout);
            this.retryTimeout += this.retryTimeoutDelta;
            return
        }
        console.log("Not retrying to connect.")
    }

    fire(e) {
        e = JSON.parse(e.data);
        console.log("->", e);
        Object.values(this.subscriptions).filter((s) => {
            s = s.event;
            // Oh boy, we need to check if the given subscription should be given the event
            if (s.event != e.event && s.event != "*") return false;
            if (s.source !== undefined && s.source != "*" && (e.source === undefined || s.source != e.source)) return false;
            if (s.connection !== undefined && s.connection != "*" && (e.connection === undefined || s.connection != e.connection)) return false;
            if (s.user !== undefined && s.user != "*" && (e.user === undefined || s.user != e.user)) return false;
            if (s.plugin !== undefined && (e.plugin === undefined || e.plugin != s.plugin)) return false;
            if (s.key !== undefined && (e.key === undefined || e.key != s.key)) return false;
            return true;
        }).forEach((s) => s.callback(e))
    }

    send(m) {
        if (this.isopen) {
            console.log("<-", m);
            this.ws.send(JSON.stringify(m));
        }
    }

    subscribe(key, event, callback) {
        this.subscriptions[key] = {
            event: event,
            callback: callback
        };
        this.send({
            cmd: "subscribe",
            ...event
        })
    }

    unsubscribe(key) {
        if (this.subscriptions[key] !== undefined) {
            let k = this.subscriptions[key]
            delete this.subscriptions[key];
            this.send({
                cmd: "unsubscribe",
                ...k.event
            });
        }
    }
}

export default EventSubscriber