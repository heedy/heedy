import api, { deepEqual } from "../../../util.mjs";
import DatasetContext, {getQueryElementTimeseries} from "./context.js";


// Get all timeseries that are included in a dataset query
const getQueryTimeseries = (q) => Object.values(q).reduce(
  (qobj, qv) => Object.assign(qobj, getQueryElementTimeseries(qv)),
  {}
);

// The Query class handles 
class Query {
    constructor(q,qJSON,userSettings,deleterFunction,getTimeseriesFunction) {
        this.query = q;
        this.queryJSON = qJSON;
        
        this._deleteMe = deleterFunction;
        this._getTimeseries = getTimeseriesFunction;
        this._userSettings = userSettings;

        // The subscriptions are all the functions to call upon data events
        this.subscriptions = new Map();

        // A promise for the currently-being-queried dataset
        this._runPromise = null;
        this._contextPromise = null;

        this._dataset = null
        this._hasNewData = false;
        this._ctx = null;

        // Once datasetPromise is finished, should it rerun the query?
        // This is used when the data is updated when the query is still active
        this._requery = false;
        this._recontext = false;

        this._abort = null;

        console.vlog(`Timeseries: Query ${this.queryJSON}`,this._getTimeseries);

        // Get the timseries objects for all the timeseries in the query.
        // The promises are completed before the context is released
        this.timeseries = getQueryTimeseries(q);
        this._timeseriesPromises = Object.keys(this.timeseries).map((tsid) => this._setTimeseries(tsid));
    }
    async _setTimeseries(tsid) {
      console.vlog("Timeseries: Getting object",tsid);
      this.timeseries[tsid] = await this._getTimeseries(tsid);
    }

    onTimeseriesChanged(tsid) {
      if (this.timeseries[tsid]!==undefined) {
        this._timeseriesPromises.push(async (tsid) => {
          await this._setTimeseries(tsid)
          if (this._contextPromise===null) {
            setTimeout(()=> this.updateContext());
          }
        });
      }
    }

    onSettingsChange(settings) {
      this._userSettings = settings;
      this.updateContext();
    }

    onDataChange(tsid) {
      if (this.timeseries[tsid]!==undefined) {
        this.run();
      }
    }

    isActive() {
        return this.subscriptions.size > 0;
    }

    subscribe(key,subs) {
        console.vlog(`Timeseries: New subscription ${key} to query ${this.queryJSON}`);
        this.subscriptions.set(key,subs);
        if (this._ctx!==null) {
          // The context is already good to go, so send it over right away
          subs.onContext(this._ctx);
        } else if (this.subscriptions.size==1) {
          // A subscription was just added, and no context is available yet.
          // This means that we should run the query in the background to prepare things.
          this.run();
        }
    }
    unsubscribe(key) {
        console.vlog("Unsubscribing",key);
        this.subscriptions.delete(key);
    }
    setStatus(status) {
        this.subscriptions.forEach((subs) => {
            if (subs.setStatus!==undefined) {
                subs.setStatus(status);
            }
        });
    }
    setError(status) {
      this.subscriptions.forEach((subs) => {
          if (subs.onError!==undefined) {
              subs.onError(status);
          }
      });
  }

    // Run the full query
    async run() {
        if (!this.isActive()) {
            return;
        }
        // The query itself is run in the _run function. This function handles the logic
        // behind querying and re-querying, and can be called multiple times while _run is being run.
        if (this._runPromise !== null) {
            console.vlog("Timeseries: waiting until current query finishes before re-processing");
            this._requery = true;
            return;
        }

        this._runPromise = this._run();

        let hadError = false;
        try {
        await this._runPromise;
        } catch (err) {
            this.setError(`Query Failed: ${err.toString()}`);
            console.error(err);
            hadError = true;
        }

        if (!this.isActive()) {
            return;
        }

        // Once the actual dataset is gathered, update its context.
        if (!hadError) {
          await this.updateContext(true);
        }

        // Then, if we're waiting to rerun the query, do so at the next opportunity.
        if (this._requery && this.isActive()) {
            setTimeout(() => {
                this._runPromise = null;
                this.run();
            });
        } else {
          this._runPromise = null;
        }

    }
    async _run() {
        console.vlog(`Timeseries: Getting dataset for ${this.queryJSON}`);
        this.setStatus("Querying Data...");
        this._requery = false;
        
        let result = await api("POST", `api/timeseries/dataset`, this.query);
        
        if (!result.response.ok) {
          throw new Error(result.data.error_description);
        }

        // Now preprocess the dataset to set all timestamps to Date objects
        Object.keys(result.data).forEach(k=> {
            result.data[k].forEach(dp=> {
                dp.t = new Date(dp.t*1000);
            });
        });

        this._dataset = result.data;
        this._hasNewData = true;
    }

    async updateContext(is_run=false) {
      
      if (!this.isActive()) {
        return;
      }
      // The query itself is run in the _run function. This function handles the logic
      // behind querying and re-querying, and can be called multiple times while _run is being run.
      if (this._contextPromise !== null) {
          console.vlog("Timeseries: waiting until current query finishes before re-processing");
          this._recontext = true;
          return;
      }
      if (this._runPromise!==null && !is_run) {
        console.vlog("Timeseries: the updated context will be used once query data is available");
        return;
      }
      

      this._contextPromise = this._updateContext();

      try {
        await this._contextPromise;
      } catch (err) {
          this.setError(`Context processing Failed: ${err.toString()}`);
          console.error(err);
      }

      // Then, if we're waiting to rerun the query, do so at the next opportunity.
      if (this._recontext && this.isActive()) {
          setTimeout(() => {
              this._contextPromise = null;
              this.updateContext();
          });
      } else {
        this._contextPromise = null;
      }

    }

    async _updateContext() {
      console.vlog(`Timeseries: Processing ${this.queryJSON}`);
      this.setStatus(`Processing Data...`);

      // Make sure that the timeseries are all there
      for (let i=0;i<this._timeseriesPromises.length;i++) {
        await this._timeseriesPromises[i];
      }
      this._timeseriesPromises = [];

      if (this._hasNewData) {
        this._hasNewData = false;
        this._ctx = new DatasetContext(this.query,this._dataset,this.timeseries,this._userSettings);
      } else if (this._ctx===null) {
          console.vlog("Timeseries: No data to process");
          return;
        } else {
          this._ctx.settings = this._userSettings;
          // The timeseries object in the context is live-updated
        }
        this.subscriptions.forEach((subs) => subs.onContext(this._ctx));
    }

}

export default Query;