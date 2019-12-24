import api from "../api.mjs";

import TimeseriesInjector from "./worker/injector.js";

import datatable from "./worker/processors/datatable.js";
import insert from "./worker/processors/insert.js";
import linechart from "./worker/processors/linechart.js";
import timeline from "./worker/processors/timeline.js";

function setup(wkr) {
  console.log("timeseries_worker: starting");

  wkr.inject("timeseries", new TimeseriesInjector(wkr));

  wkr.timeseries.addProcessor("datatable", datatable);
  wkr.timeseries.addProcessor("insert", insert);
  wkr.timeseries.addProcessor("linechart", linechart);
  wkr.timeseries.addProcessor("timeline", timeline);
}

export default setup;
