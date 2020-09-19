import Query from "./query.js";

class TimeseriesInjector {
  constructor(wkr) {
    this.worker = wkr;
    this.preprocessors = {};
    this.analyzers = [];

    // The currently active queries, keyed by their frontend key.
    this.queries = {};

    // Queries that are inactive but still contain valid data.
    this.inactive = [];

    // Subscribe to these messages from the frontend.
    // timeseries_query - a single-shot query - just return the results
    // timeseries_subscribe_query - given a key, and a query, keep the data up-to-date
    // timeseries_unsubscribe_query - unsubscribe from live updates of the given key
    wkr.addHandler("timeseries_query", (ctx, msg) => this._query(ctx, msg));
    wkr.addHandler("timeseries_subscribe_query", (ctx, msg) =>
      this._subscribeQuery(ctx, msg)
    );
    wkr.addHandler("timeseries_unsubscribe_query", (ctx, msg) =>
      this._unsubscribeQuery(ctx, msg)
    );

    // Subscribe to changes in timeseries
    wkr.websocket.subscribe(
      "timeseries_data_write",
      {
        event: "timeseries_data_write",
      },
      (e) => this._dataEvent(e)
    );
    wkr.websocket.subscribe(
      "timeseries_actions_write",
      {
        event: "timeseries_actions_write",
      },
      (e) => this._dataEvent(e)
    );
    wkr.websocket.subscribe(
      "timeseries_data_delete",
      {
        event: "timeseries_data_delete",
      },
      (e) => this._dataEvent(e)
    );

    wkr.websocket.subscribe_status((s) => this._ws_status(s));
  }
  _ws_status(s) {
    if (s === null) {
      this.inactive.forEach((q) => q.close());
      this.inactive = [];

      // Set all queries to outdated, so they are discarded on requery instead of keeping them cached
      Object.values(this.queries).forEach((q) => {
        q.outdated = true;
      });
    }
  }
  _getQuery(q, cbk, status) {
    for (let i = 0; i < this.inactive.length; i++) {
      if (this.inactive[i].isEqual(q)) {
        console.log("Using cached data for query", q);
        let qv = this.inactive[i];
        this.inactive.splice(i, 1);
        qv.activate(cbk, status);
        return qv;
      }
    }
    return new Query(this.worker, q, cbk, status);
  }
  _discardQuery(q) {
    if (this.worker.websocket.status !== null && !q.outdated) {
      // If there is an active websocket, keep the query until it no longer holds
      // up-to-date data
      console.log("Caching unused query data", q.query);
      this.inactive.push(q);
      q.deactivate(() => {
        this.inactive = this.inactive.filter((v) => v != q);
      });
    } else {
      q.close();
    }
  }
  _query(ctx, msg) {
    console.log("Running single query", msg);
    let qval = null;
    qval = this._getQuery(
      msg.query,
      (d, o) => {
        this._discardQuery(qval);
        this.worker.postMessage("timeseries_query_result", {
          key: msg.key,
          visualizations: o,
          query: d.query,
        });
      },
      (s) =>
        this.worker.postMessage("timeseries_query_status", {
          key: msg.key,
          status: s,
        })
    );
  }
  _subscribeQuery(ctx, msg) {
    console.log("Subscribing to timeseries query", msg);
    this.queries[msg.key] = this._getQuery(
      msg.query,
      (d, o) => {
        this.worker.postMessage("timeseries_query_result", {
          key: msg.key,
          visualizations: o,
          query: d.query,
        });
      },
      (s) =>
        this.worker.postMessage("timeseries_query_status", {
          key: msg.key,
          status: s,
        })
    );
  }
  _unsubscribeQuery(ctx, msg) {
    console.log("Unsubscribing from timeseries query", msg);
    if (this.queries[msg.key] !== undefined) {
      let q = this.queries[msg.key];

      delete this.queries[msg.key];

      this._discardQuery(q);
    }
  }
  _dataEvent(event) {
    console.log(
      "Data event - checking if any timeseries queries need to be updated",
      event
    );
    Object.values(this.queries).forEach((q) => q.onDataEvent(event));
    [...this.inactive].forEach((q) => q.onDataEvent(event));
  }

  /**
   * A preprocessor is an async function which is given the query, its associated dataset, as well as access to apps/timeseries
   * and the visualization settings given by an analyzer (or by person editing dashboard), and it performs any necessary preprocessing steps that might
   * take a long time/be computationally intensive. It is permitted to output a visualization of a different type than it is given.
   *
   * @param {*} vistype The visualization type to handle
   * @param {*} f An async function that performs preprocessing
   */
  addPreprocessor(vistype, f) {
    this.preprocessors[vistype] = f;
  }

  /**
   * Analyzers are async functions that given a query, its associated dataset, as well as access to apps/timeseries
   * decides which visualizations to use and how to set them up. As an example, given a numeric timeseries, an analyzer might
   * output the settings necessary to view the data as a line plot.
   *
   * @param {*} f An analysis function
   */
  addAnalyzer(f) {
    this.analyzers.push(f);
  }
}
export default TimeseriesInjector;
