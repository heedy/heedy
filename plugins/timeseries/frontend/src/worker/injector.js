import DataManager from "./datamanager.js";

class TimeseriesInjector {
  constructor(wkr) {
    this.worker = wkr;
    this.processors = {};

    this.timeseries = {};

    wkr.addHandler("timeseries_query", (ctx, msg) => this._query(ctx, msg));
    wkr.addHandler("timeseries_subscribe_query", (ctx, msg) =>
      this._subscribeQuery(ctx, msg)
    );
    wkr.addHandler("timeseries_unsubscribe_query", (ctx, msg) =>
      this._unsubscribeQuery(ctx, msg)
    );

    // TODO: In the future, make sure to subscribe to timeseries from other users
    // that might be queried
    if (wkr.info.user != null) {
      wkr.websocket.subscribe(
        "timeseries_data_write",
        {
          event: "timeseries_data_write",
          user: wkr.info.user.username
        },
        e => this._dataEvent(e)
      );
      wkr.websocket.subscribe(
        "timeseries_actions_write",
        {
          event: "timeseries_actions_write",
          user: wkr.info.user.username
        },
        e => this._dataEvent(e)
      );
      wkr.websocket.subscribe(
        "timeseries_data_delete",
        {
          event: "timeseries_data_delete",
          user: wkr.info.user.username
        },
        e => this._dataEvent(e)
      );
      /* object updates happen through re-subscribing
            wkr.websocket.subscribe("object_update_timeseries", {
                event: "object_update",
                user: wkr.info.user.username
            }, (e) => this._objectEvent(e));
            */
      wkr.websocket.subscribe(
        "object_delete_timeseries",
        {
          event: "object_delete",
          user: wkr.info.user.username
        },
        e => this._objectEvent(e)
      );

      // In a perfect world, we would also subscribe to object_update.
      // However, having the timeseries come up from the frontend instead allows
      // us to avoid an API query - otherwise each time the object is updated,
      // there would be 2 queries, one from the frontend, and one from the worker.
      // This way, the frontend queries, and the worker gets the results of that query.
      wkr.addHandler("timeseries_update", (ctx, msg) =>
        this._timeseriesUpdate(msg)
      );

      wkr.websocket.subscribe_status(s => this._ws_status(s));
    }
  }

  addProcessor(key, f) {
    this.processors[key] = f;
  }

  _ws_status(s) {
    Object.values(this.timeseries).forEach(sv => sv.onWebsocket(s));
  }

  async _dataEvent(event) {
    console.log("timeseries_worker: DATA EVENT", event);
    if (this.timeseries[event.object] !== undefined) {
      this.timeseries[event.object].onEvent(event);
    }
  }
  async _objectEvent(event) {
    console.log("timeseries_worker: object event", event);
    if (this.timeseries[event.object] !== undefined) {
      if (event.event == "object_delete") {
        this.timeseries[event.object].clear();
        delete this.timeseries[event.object];
      }
    }
  }
  async _timeseriesUpdate(timeseries) {
    if (this.timeseries[timeseries.id] !== undefined) {
      this.timeseries[timeseries.id].updateTimeseries(timeseries);
    }
  }
  async _subscribeQuery(ctx, msg) {
    if (this.timeseries[msg.timeseries.id] === undefined) {
      this.timeseries[msg.timeseries.id] = new DataManager(
        this,
        msg.timeseries
      );
    }
    this.timeseries[msg.timeseries.id].subscribe(
      msg.timeseries,
      msg.key,
      msg.query
    );
  }
  async _unsubscribeQuery(ctx, msg) {
    if (this.timeseries[msg.id] !== undefined) {
      this.timeseries[msg.id].unsubscribe(msg.key);
    }
  }
  async _query(ctx, msg) {
    if (this.timeseries[msg.timeseries.id] === undefined) {
      this.timeseries[msg.timeseries.id] = new DataManager(
        this,
        msg.timeseries
      );
    }
    this.timeseries[msg.timeseries.id].query(
      msg.timeseries,
      msg.key,
      msg.query
    );
  }
}

export default TimeseriesInjector;
