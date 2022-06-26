import TimeseriesInjector from "./worker/injector.js";

import registerAnalysis from "./worker/analysis.js";
import datatableVisualization from "./worker/visualizations/datatable.js";
import summaryVisualization from "./worker/visualizations/summary.js";
import linechartVisualization from "./worker/visualizations/linechart.js";

function setup(wkr) {
  console.vlog("timeseries_worker: starting");

  wkr.inject("timeseries", new TimeseriesInjector(wkr));

  registerAnalysis(wkr.timeseries);

  wkr.timeseries.addVisualization("datatable",datatableVisualization);
  wkr.timeseries.addVisualization("summary",summaryVisualization);
  wkr.timeseries.addVisualization("linechart",linechartVisualization);

}

export default setup;
