import api from "../rest.mjs";

import TimeseriesInjector from "./worker/injector.js";

import datatable from "./worker/processors/datatable.js";
import insert from "./worker/processors/insert.js";
import linechart from "./worker/processors/linechart.js";
import dayview from "./worker/processors/dayview.js";

function setup(wkr) {
  console.log("timeseries_worker: starting");

  wkr.inject("timeseries", new TimeseriesInjector(wkr));

  wkr.timeseries.addProcessor("datatable", datatable);
  wkr.timeseries.addProcessor("insert", insert);
  wkr.timeseries.addProcessor("linechart", linechart);
  wkr.timeseries.addProcessor("dayview", dayview);
}

export default setup;
