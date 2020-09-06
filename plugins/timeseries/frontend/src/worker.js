import TimeseriesInjector from "./worker/injector.js";

import datatableAnalyzer from "./worker/analyzers/datatable.js";
/*
import datatable from "./worker/preprocessors/datatable.js";
import insert from "./worker/preprocessors/insert.js";
import linechart from "./worker/preprocessors/linechart.js";
import dayview from "./worker/preprocessors/dayview.js";
*/
function setup(wkr) {
  console.log("timeseries_worker: starting");

  wkr.inject("timeseries", new TimeseriesInjector(wkr));

  wkr.timeseries.addAnalyzer(datatableAnalyzer);
  /*
  wkr.timeseries.addPreprocessor("datatable", datatable);
  wkr.timeseries.addPreprocessor("insert", insert);
  wkr.timeseries.addPreprocessor("linechart", linechart);
  wkr.timeseries.addPreprocessor("dayview", dayview);
  */
}

export default setup;
