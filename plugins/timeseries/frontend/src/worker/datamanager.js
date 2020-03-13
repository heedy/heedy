import api from "../../rest.mjs";
import { cleanDT } from "../analysis.mjs";

import QueryManager from "./querymanager.js";

import { objectTester } from "./util.js";
class TimeseriesDataManager {
  constructor(si, timeseries) {
    this.si = si;
    this.timeseries = timeseries;

    this.queries = {};
  }
  /**
   * This function processes heedy events to make sure that all data
   * it manages stays up-to-date
   * @param {*} e Heedy event
   */
  onEvent(e) {
    console.log("event: ", this.timeseries.id, e);
    Object.values(this.queries).forEach(q => q.onEvent(e));
  }

  onWebsocket(ws) {
    console.log("Websocket event");
    Object.values(this.queries).forEach(q => q.onWebsocket(ws));
  }

  /**
   * Subscribes to the given query. The results of the query are kept up-to-date
   * @param {*} key
   * @param {*} query
   */
  subscribe(timeseries, key, query) {
    console.log(
      this.timeseries.id,
      "timeseries_worker:  subscribe ",
      key,
      query
    );
    this.updateTimeseries(timeseries);
    this.queries[key] = new QueryManager(
      q => this.runquery(q),
      query,
      d => this.process_and_send(key, d)
    );
  }

  clear() {
    Object.values(this.queries).forEach(q => q.close());
    this.queries = {};
  }

  /**
   * Unsubscribes from the given query, meaning that its results will no longer
   * be kept up-to-date
   * @param {*} key
   */
  unsubscribe(key) {
    console.log(this.timeseries.id, "timeseries_worker:  unsubscribe ", key);
    delete this.queries[key];
  }

  /**
   * Updates timeseries metadata
   * @param {*} timeseries
   */
  updateTimeseries(timeseries) {
    if (!objectTester(timeseries, this.timeseries)) {
      console.log("timeseries_worker: timeseries updated", timeseries);
      this.timeseries = timeseries;
      this.refresh();
    }
  }

  async recompute(key) {
    let d = await this.queries[key].data();
    await this.process_and_send(key, d);
  }

  refresh() {
    Object.keys(this.queries).forEach(k => this.recompute(k));
  }

  async runquery(query) {
    let result = await api(
      "GET",
      `api/objects/${this.timeseries.id}/timeseries`,
      query
    );
    if (!result.response.ok) {
      throw result.response.error_message;
    }
    return result.data;
  }

  async process(data) {
    let vals = Object.values(this.si.processors).map(v =>
      v(this.timeseries, cleanDT(data))
    );
    let outvals = {};
    for (let i = 0; i < vals.length; i++) {
      let res = await vals[i];
      outvals = Object.assign(outvals, res);
    }
    return outvals;
  }

  async process_and_send(key, data) {
    this.si.worker.postMessage("timeseries_views", {
      key,
      id: this.timeseries.id,
      views: {
        query_status: {
          view: "status",
          data: `Processing ${data.length.toLocaleString()} Datapoints`
        }
      }
    });
    let outvals = await this.process(data);
    this.si.worker.postMessage("timeseries_views", {
      key,
      id: this.timeseries.id,
      views: outvals
    });
  }

  async query(timeseries, key, query, qcallback = x => x) {
    console.log("timeseries_worker: Querying ", this.timeseries.id, key, query);

    try {
      var data = await this.runquery(query);
    } catch (err) {
      this.si.worker.postMessage("timeseries_views", {
        key,
        id: this.timeseries.id,
        views: {
          error: {
            view: "error",
            data: result.response.error_message
          }
        }
      });
      return;
    }
    console.log("timeseries_worker: Queried ", data.length);
    await this.process_and_send(key, data);
  }
}

export default TimeseriesDataManager;
