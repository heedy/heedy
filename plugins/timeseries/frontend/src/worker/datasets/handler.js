import Query from "./query.js";
import {
  stableStringify
} from "../../../dist/json-json-template.mjs";
import {
  preprocessAll
} from "./preprocessConfig.js";
import {
  deepEqual
} from "../../../util.mjs";

// User visualizations are javascript code. Since the code will be re-imported each time the websocket
// reconnects, we want to cache the compiled objects to avoid recompiling.
const userVisualizationCache = new Map();

// Compile the user visualizations, and cache the result if cache is true.
const prepareUserVisualizations = (vis, cache = true) => vis.filter(v => v.enabled).map(v => {
  let outf = userVisualizationCache.get(v.code);
  if (outf === undefined) {
    console.vlog("Compiling user visualization", v.name);
    try {
      outf = new Function("c", "vis", v.code);
      if (cache) {
        userVisualizationCache.set(v.code, outf);
      }
    } catch (e) {
      // The function failed to compile, so replace it with a function that rethrows the error.
      // Errors thrown in visualization functions will be caught, and returned to the user.
      e.message = `Failed to compile user visualization '${v.name}': ${e.message}`;
      outf = (c) => {
        throw e;
      };
    }
  }

  return {
    name: v.name,
    f: outf,
    code: v.code
  };
});

// Given a query context, and visualizations arrays, generates the visualization object.
const getVisualizations = (c, visualizations, user_visualizations) => {
  let vis = {};
  let errors = [];
  for (let v of visualizations) {
    try {
      const vis2 = v.f(c, vis);
      if (vis2 !== undefined) {
        vis = vis2;
      }
    } catch (e) {
      errors.push({
        type: "plugin",
        error: e.toString(),
        name: v.name
      });
    }
  }

  for (let v of user_visualizations) {
    try {
      const vis2 = v.f(c, vis);
      if (vis2 !== undefined) {
        vis = vis2;
      }
    } catch (e) {
      errors.push({
        type: "user",
        error: e.toString(),
        name: v.name
      });
    }
  }

  if (errors.length > 0) {
    vis.errors = {
      type: "visualization_errors",
      title: "Visualization Errors",
      weight: -100,
      config: {
        errors
      }
    };
  }
  return vis;
}


class DatasetQueryHandler {
  constructor(wkr) {

    // The currently active subscriptions, keyed by their frontend key. The map holds the
    // query objects for each subscription. A single query object can be used by multiple subscriptions.
    this.subscriptions = new Map();

    // The currently up-to-date queries. Keyed by the json of the query.
    // When a query is no longer subscribed, it stays in this map until its data
    // is outdated, at which point it removes itself.
    this.queries = new Map();

    // List of visualizations, both built-in and added by plugins
    this.visualizations = [];
    // List of the user's custom visualizations.
    this.user_visualizations = prepareUserVisualizations(wkr.info.settings.timeseries.visualizations);

    this.settings = wkr.info.settings;

    // Subscribe to these messages from the frontend.
    // timeseries_query - a single-shot query - just return the results
    // timeseries_subscribe_query - given a key, and a query, keep the data up-to-date
    // timeseries_unsubscribe_query - unsubscribe from live updates of the given key
    wkr.addHandler("timeseries_query", (ctx, msg) => console.log(ctx, msg));
    wkr.addHandler("timeseries_subscribe_query", (ctx, msg) => this.subscribe(msg));
    wkr.addHandler("timeseries_unsubscribe_query", (ctx, msg) => this.unsubscribe(msg.key));

    // Next, subscribe to data events and websocket events to keep things up-to-date.
    wkr.websocket.subscribe(
      "timeseries_data_write", {
        event: "timeseries_data_write",
      },
      (e) => this._dataChanged(e)
    );
    wkr.websocket.subscribe(
      "timeseries_data_delete", {
        event: "timeseries_data_delete",
      },
      (e) => this._dataChanged(e)
    );

    // And finally, subscribe to info change events,
    // so that we can update the user visualizations
    wkr.onUserSettingsChanged((s) => this.updateSettings(s))

    wkr.websocket.subscribe_status((s) => this._websocketConnectionStatusChanged(s));

    this.worker = wkr;

  }

  // Whenever the user visualizations are updated, we need to recompile them
  updateSettings(settings) {
    console.vlog("DS: Updated settings");
    const v = settings.timeseries.visualizations;
    this.user_visualizations = prepareUserVisualizations(v);

    // Next, all queries need their settings updated.
    for (let q of this.queries.values()) {
      q.onSettingsChange(settings);
    }

  }

  // When the frontend websocket is disconnected, we label all queries as outdated,
  // since we are no longer notified of updates.
  _websocketConnectionStatusChanged(s) {
    console.vlog("DS: WEBSOCKET STATUS", s);
    if (s!==null) {
      for (let q of this.queries.values()) {
        q.run();
      }
    }

  }

  _dataChanged(event) {
    console.vlog("DATA CHANGE",event);
    for (let q of this.queries.values()) {
      q.onDataChange(event.object);
    }
  }


  addVisualization(v) {
    this.visualizations.push(v);
  }

  unsubscribe(key) {
    const q = this.subscriptions.get(key);
    if (q !== undefined) {
      this.subscriptions.delete(key);
      q.unsubscribe(key);
    }
  }

  // Subscribe to a query, keeping it active and updated
  subscribe(msg) {
    // Generate a subscription for the query
    const setStatus = (status) => this.worker.postMessage("timeseries_query_status", {
      key: msg.key,
      status: status,
    });

    // See if we use the user visualizations, or custom ones
    let user_visualizations = null;
    if (msg.user_visualizations !== undefined && msg.user_visualizations !== null) {
      user_visualizations = prepareUserVisualizations(msg.user_visualizations, false);
    }

    const onContext = (c) => {
      console.vlog("DS: New vis context", c);

      // Get the visualizations
      let vis = getVisualizations(c, this.visualizations, user_visualizations != null ? user_visualizations : this.user_visualizations);

      // Then preprocess them
      const out = preprocessAll(c, vis);

      this.worker.postMessage("timeseries_query_result", {
        key: msg.key,
        visualizations: out,
        query: msg.query,
      });
    }

    const subscription = {
      setStatus,
      onContext,
      onError: setStatus
    };

    // Check if there is a cached Query object valid for this subscription
    const queryJSON = stableStringify(msg.query);
    let q = this.queries.get(queryJSON);
    if (q === undefined) {
      // No cached query, so create a new one, telling it to remove itself
      // when it is no longer required and is outdated. This will only be called
      // once there are no active subscriptions for the query.
      q = new Query(msg.query, queryJSON, this.settings, () => this.queries.delete(queryJSON), (tsid) => this.worker.objects.get(tsid));
      this.queries.set(queryJSON, q);
    }
    q.subscribe(msg.key, subscription);
    this.subscriptions.set(msg.key, q);
  }

  // Run a query, but don't keep it active
  run(msg) {
    // This is identical to subscribe, but it doesn't keep the query active,
    // meaning that the query is unsubscribed when it returns a result or error.
  }
}

export default DatasetQueryHandler;