/**
 * The QueryManager gets a function that can be queried to get data for the given stream,
 * the query to keep track of, and a callback that is given the actual data as an argument,
 * and which is called whenever the data is actually changed
 */
class QueryManager {
    constructor(runquery, query, datacallback) {
        this.runquery = runquery;
        this.query = query;
        this.datacallback = datacallback;
        this.d = null;
        this.querying = runquery(query).then(d => {
            this.d = d;
            this.querying = null;
            datacallback(d);
        });
    }

    onEvent(e) {
        console.log("Query event:", e);
        this.querying = this.runquery(this.query).then(d => {
            this.d = d;
            this.querying = null;
            this.datacallback(d);
        });
    }

    onWebsocket(ws) {
        console.log("Query websocket", ws);
    }

    async data() {
        if (this.querying == null) {
            return this.d;
        }
        await this.querying;
        return this.d;
    }

    close() {
        if (this.querying != null) {
            this.querying.cancel();
        }
    }

}

export default QueryManager;