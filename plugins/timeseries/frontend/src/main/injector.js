class TimeseriesInjector {
  constructor(app) {
    this.app = app;

    this.subscriptions = {};

    app.worker.addHandler("timeseries_views", (c, m) => this._onViews(c, m));

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
      console.error("Unknown timeseries view subscription key ", skey);
      return;
    }
    this.subscriptions[skey](msg.views);
  }

  subscribeQuery(timeseries, key, query, callback) {
    let skey = timeseries.id + ":" + key;
    this.subscriptions[skey] = callback;
    this.app.worker.postMessage("timeseries_subscribe_query", {
      timeseries: timeseries,
      key,
      query
    });
    if (this.watchers[timeseries.id] === undefined) {
      this.watchers[timeseries.id] = this.app.store.watch(
        (state, getters) => state.heedy.objects[timeseries.id],
        (n, o) => {
          if (n === undefined || n === null) {
            console.log("Stopping watch of ", timeseries.id);
            this.watchers[timeseries.id]();
            return;
          }
          this.app.worker.postMessage("timeseries_update", n);
        }
      );
    }
  }
  unsubscribeQuery(tsid, key) {
    this.app.worker.postMessage("timeseries_unsubscribe_query", {
      id: tsid,
      key
    });
    let skey = tsid + ":" + key;
    delete this.subscriptions[skey];
  }

  query(timeseries, query, callback) {
    let skey = timeseries.id + ":" + key;
    this.subscriptions[skey] = d => {
      delete this.subscriptions[skey];
      callback(d);
    };
    this.app.worker.postMessage("timeseries_query", {
      timeseries: timeseries,
      key,
      query
    });
  }
}

export default TimeseriesInjector;
