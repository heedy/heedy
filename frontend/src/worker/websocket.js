import moment from "../dist/moment.mjs";

class WebsocketInjector {
  constructor(wkr) {
    this.wkr = wkr;

    this.subscriptions = {};

    // The status of the websocket
    this.status = null;
    this.status_callbacks = [];

    wkr.addHandler("websocket_event", (ctx, msg) =>
      this.subscriptions[msg.key](msg.event)
    );
    wkr.addHandler("websocket_status", (ctx, msg) => {
      if (msg != null) {
        msg = moment(msg);
      }

      this.status = msg;
      this.status_callbacks.forEach((c) => c(msg));
    });
  }

  subscribe(key, event, callback) {
    this.subscriptions[key] = callback;
    this.wkr.postMessage("websocket_subscribe", {
      key,
      event,
    });
  }

  unsubscribe(key) {
    this.wkr.postMessage("websocket_unsubscribe", {
      key,
    });
    delete this.subscriptions[key];
  }

  subscribe_status(callback) {
    this.status_callbacks.push(callback);
  }
}

export default WebsocketInjector;
