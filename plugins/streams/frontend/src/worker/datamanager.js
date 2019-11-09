import api from "../../api.mjs";

import QueryManager from "./querymanager.js";

import {
    objectTester
} from "./util.js";
class StreamDataManager {
    constructor(si, stream) {
        this.si = si;
        this.stream = stream;

        this.queries = {}

    }
    /**
     * This function processes heedy events to make sure that all data 
     * it manages stays up-to-date
     * @param {*} e Heedy event 
     */
    onEvent(e) {
        console.log("event: ", this.stream.id, e);
        Object.values(this.queries).forEach(q => q.onEvent(e));
    }

    onWebsocket(ws) {
        console.log("Websocket event");
        Object.values(this.queries).forEach(q => q.onWebsocket(ws));
    }

    /**
     * Subscribes to the given query. The results of the query are kept up-to-date
     * @param {*} key 
     * @param {*} query 
     */
    subscribe(stream, key, query) {
        console.log(this.stream.id, "stream_worker:  subscribe ", key, query);
        this.updateStream(stream);
        this.queries[key] = new QueryManager((q) => this.runquery(q), query, (d) => this.process_and_send(key, d));
    }

    clear() {
        Object.values(this.queries).forEach(q => q.close());
        this.queries = {};
    }


    /**
     * Unsubscribes from the given query, meaning that its results will no longer 
     * be kept up-to-date
     * @param {*} key 
     */
    unsubscribe(key) {
        console.log(this.stream.id, "stream_worker:  unsubscribe ", key);
        delete this.queries[key];
    }

    /**
     * Updates stream metadata
     * @param {*} stream 
     */
    updateStream(stream) {
        if (!objectTester(stream, this.stream)) {
            console.log("stream_worker: stream updated", stream);
            this.stream = stream;
            this.refresh();
        }

    }

    async recompute(key) {
        let d = await this.queries[key].data();
        await this.process_and_send(key, d);
    }


    refresh() {
        Object.keys(this.queries).forEach(k => this.recompute(k));
    }

    async runquery(query) {
        let result = await api("GET", `api/heedy/v1/sources/${this.stream.id}/data`, query);
        if (!result.response.ok) {
            throw result.response.error_message;
        }
        return result.data;
    }

    async process(data) {
        let vals = Object.values(this.si.processors).map(v => v(this.stream, data));
        let outvals = {};
        for (let i = 0; i < vals.length; i++) {
            let res = await vals[i];
            outvals = Object.assign(outvals, res);
        }
        return outvals;
    }

    async process_and_send(key, data) {
        let outvals = await this.process(data);
        this.si.worker.postMessage("stream_views", {
            key,
            id: this.stream.id,
            views: outvals
        });
    }

    async query(stream, key, query) {
        console.log("stream_worker: Querying ", this.stream.id, key, query);

        try {
            var data = await this.runquery(query);
        } catch (err) {
            this.si.worker.postMessage("stream_views", {
                key,
                id: this.stream.id,
                views: {
                    error: {
                        view: "error",
                        data: result.response.error_message
                    }
                }
            });
        }
        console.log("stream_worker: Query Result", data);
        await this.process_and_send(key, data);
    }
}

export default StreamDataManager;