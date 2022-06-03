
var randomKey = () =>
  "_" +
  Math.random()
    .toString(36)
    .substr(2, 9);

/**
 * @alias frontend.timeseries
 */
class TimeseriesInjector {
  /**
   * Hello
   * @param {*} app 
   */
  constructor(app) {
    this.app = app;

    this.subscriptions = {};

    app.worker.addHandler("timeseries_query_result", (c, m) =>
      this._onQueryResult(c, m)
    );
    app.worker.addHandler("timeseries_query_status", (c, m) =>
      this._onQueryResult(c, m)
    );
  }
  /**
   * Add a visualization display component for a visualization type
   * @param {*} name - the name of the visualization type
   * @param {*} component 
   */
  addVisualizationType(name, component) {
    this.app.store.commit("addTSVisualizationType", {
      key: name,
      component: component,
    });
  }

  addType(value) {
    this.app.store.commit("addTSType", value);
  }

  _onQueryResult(ctx, msg) {
    if (this.subscriptions[msg.key] === undefined) {
      console.error("Unknown timeseries query subscription key ", msg.key);
      return;
    }
    this.subscriptions[msg.key](msg);
  }

  subscribeQuery(query, callback,user_visualizations=null) {
    let key = randomKey();
    this.subscriptions[key] = callback;
    this.app.worker.postMessage("timeseries_subscribe_query", {
      key,
      query,
      user_visualizations,
    });
    return key;
  }
  unsubscribeQuery(key) {
    this.app.worker.postMessage("timeseries_unsubscribe_query", {
      key,
    });
    delete this.subscriptions[key];
  }

  query(q) {
    return new Promise((resolve, reject) => {
      let key = randomKey();
      this.subscriptions[key] = (d) => {
        delete this.subscriptions[key];
        resolve(d);
      };
      this.app.worker.postMessage("timeseries_query", {
        key,
        query,
      });
    });
  }
}

export default TimeseriesInjector;
