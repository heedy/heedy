import WebsocketInjector from "./worker/websocket.js";
import ObjectInjector from "./worker/objects.js";

class Wrkr {
  constructor() {

    // app info goes here
    this.info = null;

    // Start out not logging until we get info message to determine whether
    // logging is OK
    if (console.vlog === undefined) {
      console.vlog = (a, b) => { }
    }


    this.handlers = {
      import: (ctx, data) => this._importHandler(ctx, data),
      info: (ctx, data) => {
        this.info = data;

        // The v-functions are logging functions that are conditional on whether
        // the server is in verbose mode
        if (!_DEBUG && !data.verbose) {
          let c = (a, b, d, e, f) => { };
          console.vdebug = c;
          console.vlog = c;
          console.vwarn = c;
          console.verror = c;
          console.vinfo = c;
          console.vtable = c;
        } else {
          console.vdebug = console.debug;
          console.vlog = console.log;
          console.vwarn = console.warn;
          console.verror = console.error;
          console.vinfo = console.info;
          console.vtable = console.table;
        }

        console.vlog("worker: started");
      },
    };

    // The worker needs to enforce an import ordering, because a message
    // might use stuff from a just-imported worker. Therefore, we keep a
    // queue of messages that need to be executed in order
    this.messageQueue = [];

    this.inject("websocket", new WebsocketInjector(this));
    this.inject("objects", new ObjectInjector(this));
  }

  addHandler(key, f) {
    this.handlers[key] = f;
  }

  async _importHandler(ctx, msg) {
    try {
      (await msg).default(this);
    } catch (err) {
      console.error(err);
    }
  }
  postMessage(key, msg) {
    // post message
    postMessage({
      key: key,
      msg: msg,
    });
  }
  async queueHandler() {
    while (this.messageQueue.length > 0) {
      let msg = this.messageQueue[0];

      console.vlog("worker: processing ", msg);
      if (this.handlers[msg.key] !== undefined) {
        let ctx = {
          key: msg.key,
        };
        await this.handlers[msg.key](ctx, msg.msg);
      } else {
        console.error(`worker: unknown handler ${msg.key}`);
      }
      this.messageQueue.shift();
    }
  }

  _onMessage(e) {
    let msg = e.data;
    console.vlog("worker: received ", msg);

    // We use special handling for import messages, so that they start loading right away
    if (msg.key == "import") {
      console.vlog("worker: import", msg.msg);
      msg.msg = import("./" + msg.msg);
    }
    this.messageQueue.push(msg);
    if (this.messageQueue.length > 1) {
      return;
    }
    this.queueHandler();
  }

  inject(name, p) {
    this[name] = p;
  }
}

let worker = new Wrkr();

// In the future this won't be necessary, since this will be a worker,
// but for now, we just emulate the worker.
export default worker;
