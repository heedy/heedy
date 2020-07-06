class DashboardInjector {
  constructor(app) {
    this.app = app;
    this.subscriptions = {};

    app.worker.addHandler("dashboard_update", (c, m) => this._onUpdate(c, m));
  }

  _onUpdate(ctx, msg) {
    if (this.subscriptions[msg.id] === undefined) {
      console.warn("No dashboard subscription for ", msg.id);
      return;
    }
    this.subscriptions[msg.id].forEach((f) => f.callback(msg.data));
  }

  subscribe(id, key, callback) {
    if (this.subscriptions[id] === undefined) {
      this.subscriptions[id] = [];
    }
    this.subscriptions[id].push({ callback: callback, key: key });
    this.app.worker.postMessage("dashboard_subscribe", { id: id });
  }
  unsubscribe(id, key) {
    this.app.worker.postMessage("dashboard_unsubscribe", { id: id });
    if (this.subscriptions[id] === undefined) {
      return;
    }
    this.subscriptions[id] = this.subscriptions[id].filter((e) => e.key != key);
    if (this.subscriptions[id].length == 0) {
      delete this.subscriptions[id];
    }
  }

  addType(type, component) {
    this.app.store.commit("addDashboardType", {
      type: type,
      component: component,
    });
  }
}

export default DashboardInjector;
