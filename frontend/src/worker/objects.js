class ObjectInjector {
  constructor(wkr) {
    this.wkr = wkr;

    // The cached objects
    this.objects = {};

    // Callbacks waiting for an object
    this.waiting = {};

    // Subscriptions to each object
    this.subscriptions = {};

    this.ws_initialized = false;

    wkr.websocket.subscribe_status((s) => this._ws_status(s));

    wkr.addHandler("get_object", (ctx, msg) => {
      this.objects[msg.id] = msg;
      if (this.waiting[msg.id] !== undefined) {
        this.waiting[msg.id].forEach((c) => c(msg));
        delete this.waiting[msg.id];
      }
    });
  }

  _ws_status(evt) {
    console.log("WS STATUS:", evt);
    if (evt !== null && !this.ws_initialized) {
      // This subscribe needs to be deferred until at least the info message is sent,
      // since no guarantee frontend is listening until then.
      // TODO: Once heedy supports it, subscribe to objects from other users too?
      this.wkr.websocket.subscribe(
        "worker_object_deleted",
        {
          event: "object_delete",
        },
        (e) => this._objectDeleted(e)
      );
      this.wkr.websocket.subscribe(
        "worker_object_updated",
        {
          event: "object_update",
        },
        (e) => this._objectUpdated(e)
      );
      this.ws_initialized = true;
      return;
    }

    if (evt == null) {
      this.objects = {};
    }
  }

  _objectDeleted(evt) {
    delete this.objects[evt.object];
    if (this.subscriptions[evt.object] !== undefined) {
      this.subscriptions[evt.object].forEach((s) => s(null));
      delete this.subscriptions[evt.object];
    }
  }
  _objectUpdated(evt) {
    delete this.objects[evt.object];
    if (this.subscriptions[evt.object] !== undefined) {
      this.get(evt.object).then((o) => {
        if (this.subscriptions[evt.object] !== undefined) {
          this.subscriptions[evt.object].forEach((s) => s(o));
        }
      });
    }
  }

  /**
   * Returns a promise for the object of the given ID
   * @param {string} id The object ID
   */
  get(id) {
    return new Promise((resolve, reject) => {
      if (this.objects[id] !== undefined) {
        resolve(this.objects[id]);
        return;
      }
      if (this.waiting[id] === undefined) {
        this.waiting[id] = [resolve];
        this.wkr.postMessage("get_object", { id: id });
      } else {
        this.waiting[id].push(resolve);
      }
    });
  }

  /**
   * Called on object update and delete. Also called on websocket reconnect.
   *
   * @param {*} id The ID of the object
   * @param {*} callback a callback to use whenever the object is changed. It is passed null if the object is deleted.
   */
  subscribe(id, callback) {
    if (this.subscriptions[id] === undefined) {
      this.subscriptions[id] = [callback];
    } else {
      this.subscriptions[id].push(callback);
    }
    // Returns a subscription object
    return {
      id: id,
      callback: callback,
    };
  }

  /**
   * Remove an existing subscription to an object.
   *
   * @param {*} subscription - returned by subscribe, an object containing the callback and object ID
   */
  unsubscribe(subscription) {
    if (this.subscriptions[subscription.id] !== undefined) {
      subs = this.subscriptions[subscription.id];
      for (let i = subs.length; i >= 0; i--) {
        if (subs[i] === subscription.callback) {
          subs.splice(i, 1);
        }
      }
      if (subs.length == 0) {
        delete this.subscriptions[subscription.id];
      }
    }
  }
}

export default ObjectInjector;
