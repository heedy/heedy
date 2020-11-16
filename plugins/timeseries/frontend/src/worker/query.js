import api from "../../util.mjs";

import QueryData from "./query_data.js";

import { deepEqual } from "../../util.mjs";

function getQueryElementObjects(elem) {
  let qobj = {};
  if (elem.timeseries !== undefined) {
    qobj[elem.timeseries] = true;
  }
  if (elem.merge !== undefined) {
    elem.merge.forEach((e2) => {
      qobj = {
        ...qobj,
        ...getQueryElementObjects(e2),
      };
    });
  }
  if (elem.dataset !== undefined) {
    Object.values(elem.dataset).forEach((e2) => {
      qobj = {
        ...qobj,
        ...getQueryElementObjects(e2),
      };
    });
  }
  return qobj;
}

function getQueryObjects(q) {
  return q.reduce(
    (qobj, qv) => ({ ...qobj, ...getQueryElementObjects(qv) }),
    {}
  );
}

class Query {
  constructor(wkr, q, cbk, status) {
    this.worker = wkr;
    this.active = true;
    this.query = q;
    this.objects = getQueryObjects(q);
    this.deactivator = null;
    this.onOutput = null;
    this.onStatus = null;
    this.outdated = false;

    this.qdata = null; // Raw result of query
    this.output = null; // The output of the query
    console.log(
      "Query",
      q,
      "includes the following timeseries:",
      Object.keys(this.objects)
    );

    this.object_subscriptions = {};

    this.querypromise = null;
    this.requery = false;
    this.reprepare = false;

    this.activate(cbk, status);
  }

  activate(cbk, status) {
    this.onOutput = cbk;
    this.onStatus = status;
    this.deactivator = null;
    this.outdated = this.worker.websocket.status === null;
    if (this.qdata === null) {
      if (!this.requery) {
        this.runquery();
      }
    } else if (this.output === null) {
      // Must recompute output from the raw data
      this.prepareOutput();
    } else {
      cbk(this.qdata, this.output);
    }
  }
  /**
   * Deactivates the query, and calls the callback once its cached data is no longer up-to-date
   * @param {*} callback
   */
  deactivate(callback) {
    this.deactivator = callback;
    this.requery = false;
    this.reprepare = false;
    this.onOutput = null;
  }
  isEqual(other) {
    return deepEqual(this.query, other);
  }
  close() {
    console.log("DatasetQuery: Closing", this.query);
    Object.values(this.object_subscriptions).forEach((v) =>
      this.worker.objects.unsubscribe(v)
    );

    this.requery = false;
  }

  onDataEvent(evt) {
    if (this.objects[evt.object] !== undefined) {
      if (this.deactivator !== null) {
        console.log(
          "DatasetQuery: Query output might have changed. Removing from cache.",
          evt,
          this.query
        );
        this.deactivator();
        this.close();
      } else {
        // Still active - rerun query
        console.log(
          "DatasetQuery: Query output might have changed. Rerunning query.",
          evt,
          this.query
        );
        this.runquery();
      }
    }
  }
  onObjectEvent(obj) {
    console.log("OBJECT EVENT", obj);
    if (this.deactivator === null) {
      this.prepareOutput();
    } else {
      this.output = null; // The output is no longer valid
    }
  }

  async runquery() {
    if (this.querypromise != null) {
      console.log(
        "DatasetQuery: waiting until current query finishes before re-querying"
      );
      this.requery = true;
      return;
    }
    this.querypromise = this._runquery();

    try {
      await this.querypromise;
    } catch (err) {
      if (this.deactivator === null) {
        this.onStatus(`Query Failed: ${err.toString()}`);
      }
      console.error(err);
    }
    this.querypromise = null;
    if (this.deactivator === null) {
      if (this.requery) {
        // Run another query in setTimeout to avoid recursion limit
        setTimeout(() => {
          this.requery = false;
          this.runquery();
        });
        return;
      }
      if (this.reprepare) {
        setTimeout(() => {
          this.reprepare = false;
          this.prepareOutput();
        });
      }
    }
  }
  async prepareOutput() {
    if (this.querypromise != null) {
      console.log(
        "DatasetQuery: waiting until current query finishes before re-processing"
      );
      this.reprepare = true;
      return;
    }
    this.querypromise = this._prepareOutput();

    try {
      await this.querypromise;
    } catch (err) {
      if (this.deactivator === null) {
        this.onStatus(`Processing Failed: ${err.toString()}`);
      }
      console.error(err);
    }
    this.querypromise = null;
    if (this.deactivator === null) {
      if (this.requery) {
        // Run another query in setTimeout to avoid recursion limit
        setTimeout(() => {
          this.requery = false;
          this.runquery();
        });
        return;
      }
      if (this.reprepare) {
        setTimeout(() => {
          this.reprepare = false;
          this.prepareOutput();
        });
      }
    }
  }

  async _runquery(q) {
    console.log("DatasetQuery: Getting dataset for query", this.query);
    this.onStatus("Querying Data...");
    this.requery = false;
    let result = await api("POST", `api/timeseries/dataset`, this.query);
    if (!result.response.ok) {
      throw result.data.error_description;
    }

    // If we just got data, and the websocket is on, we are not outdated
    if (this.worker.websocket.status !== null) {
      this.outdated = false;
    }

    this.qdata = new QueryData(this.worker, this, result.data);

    if (this.deactivator !== null) {
      this.deactivator();
      this.close();
      return;
    }

    await this._prepareOutput();
  }

  async _prepareOutput() {
    console.log("DatasetQuery: Processing data...", this.query);
    this.onStatus("Processing Data...");
    this.reprepare = false;
    // Start running all the analyzers
    let aresults = this.worker.timeseries.analyzers.map((a) => a(this.qdata));

    // Wait for the results
    let resultarray = await Promise.all(aresults);

    // Reduce results into a single object
    let results = resultarray.reduce((c, v) => Object.assign(c, v), {});

    if (this.deactivator !== null) {
      this.deactivator();
      this.close();
      return;
    }

    // Now run all preprocessors
    let rkeys = Object.keys(results);
    let rvalues = await Promise.all(
      rkeys.map((k) => {
        if (
          this.worker.timeseries.preprocessors[results[k].visualization] !==
          undefined
        ) {
          return this.worker.timeseries.preprocessors[results[k].visualization](
            this.qdata,
            results[k]
          );
        }
        return results[k];
      })
    );

    this.output = rkeys.reduce((o, cv, i) => {
      o[cv] = rvalues[i];
      return o;
    }, {});

    // Call the callback
    if (this.onOutput !== null) {
      this.onOutput(this.qdata, this.output);
    }
  }
}

export default Query;
