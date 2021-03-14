import api from "../../util.mjs";

class DashboardWorkerInjector {
  constructor(wkr) {
    this.worker = wkr;

    this.handlers = {};
    this.subscribed = {};

    wkr.addHandler("dashboard_subscribe", (ctx, msg) =>
      this._subscribe(ctx, msg)
    );
    wkr.addHandler("dashboard_unsubscribe", (ctx, msg) =>
      this._unsubscribe(ctx, msg)
    );

    // TODO: In the future, make sure to subscribe to dashboards from other users
    // that might be queried
    if (wkr.info.user != null) {
      wkr.websocket.subscribe(
        "dashboard_element_delete",
        {
          event: "dashboard_element_delete",
          user: wkr.info.user.username,
        },
        (e) => this._event(e)
      );
      wkr.websocket.subscribe(
        "dashboard_element_update",
        {
          event: "dashboard_element_update",
          user: wkr.info.user.username,
        },
        (e) => this._event(e)
      );
      wkr.websocket.subscribe(
        "dashboard_element_create",
        {
          event: "dashboard_element_create",
          user: wkr.info.user.username,
        },
        (e) => this._event(e)
      );
    }
  }

  _subscribe(ctx, msg) {
    console.vlog("dashboard_worker: subscribe", msg.id);
    if (this.subscribed[msg.id] === undefined) {
      this.subscribed[msg.id] = 0;
    }
    this.subscribed[msg.id] += 1;

    // Run a full query
    this._getDashboard(msg.id);
  }
  _unsubscribe(ctx, msg) {
    console.vlog("dashboard_worker: unsubscribe", msg.id);
    this.subscribed[msg.id] -= 1;
    if (this.subscribed[msg.id] <= 0) {
      delete this.subscribed[msg.id];
    }
  }

  _event(e) {
    if (this.subscribed[e.object] !== undefined) {
      if (e.event == "dashboard_element_delete") {
        this.worker.postMessage("dashboard_update", {
          id: e.object,
          data: [{ delete: e.data.element_id }],
        });
        return;
      }
      this._getDashboardElement(e.object, e.data.element_id);
    }
  }

  _preprocess(e) {
    if (this.handlers[e.type] !== undefined) {
      e = this.handlers[e.type](e);
    }
    return e;
  }

  async _fullquery(id) {
    let result = await api("GET", `api/objects/${encodeURIComponent(id)}/dashboard`);
    if (!result.response.ok) {
      throw result.response.error_message;
    }
    return result.data;
  }
  async _query(id, eid) {
    let result = await api("GET", `api/objects/${encodeURIComponent(id)}/dashboard/${encodeURIComponent(eid)}`);
    if (!result.response.ok) {
      throw result.response.error_message;
    }
    return result.data;
  }

  async _getDashboard(id) {
    let res = await this._fullquery(id);
    if (this.subscribed[id] !== undefined) {
      this.worker.postMessage("dashboard_update", {
        id: id,
        data: res.map((e) => this._preprocess(e)),
      });
    }
  }
  async _getDashboardElement(id, eid) {
    let res = await this._query(id, eid);
    if (this.subscribed[id] !== undefined) {
      this.worker.postMessage("dashboard_update", {
        id: id,
        data: [this._preprocess(res)],
      });
    }
  }

  preprocessType(t, f) {
    this.handlers[t] = f;
  }
}

export default DashboardWorkerInjector;
