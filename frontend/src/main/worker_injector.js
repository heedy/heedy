import worker from "../worker.mjs";

/**
 * @alias frontend.worker
 */
class WorkerInjector {
  /**
   * The worker
   * @param {*} appinfo
   */
  constructor(appinfo) {
    this.handlers = {};

    worker.postMessage = (key, msg) => {
      return this._onMessage({
        data: {
          key: key,
          msg: msg,
        },
      });
    };

    this.worker = {
      postMessage: (msg) =>
        worker._onMessage({
          data: msg,
        }),
    };

    // post the info
    this.postMessage("info", appinfo);
  }
  addHandler(key, f) {
    this.handlers[key] = f;
  }

  /**
   * Sends a message with the given key to the worker
   * @param {*} key
   * @param {*} msg
   */
  postMessage(key, msg) {
    this.worker.postMessage({
      key: key,
      msg: msg,
    });
  }

  import(filename) {
    this.postMessage("import", filename);
  }

  async _onMessage(e) {
    let msg = e.data;
    console.vlog("from_worker:", msg);
    if (this.handlers[msg.key] !== undefined) {
      let ctx = {
        key: msg.key,
      };
      await this.handlers[msg.key](ctx, msg.msg);
    } else {
      console.error(`Unknown message key ${msg.key}`);
    }
  }
}

export default WorkerInjector;
