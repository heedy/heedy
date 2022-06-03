import DatasetHandler from "./datasets/handler.js";
import { addAnalysisFunction, initDatasetContext } from "./datasets/context.js";
import {customPreprocessors} from "./datasets/preprocessConfig.js";

class TimeseriesInjector {
  constructor(wkr) {
    this.worker = wkr;

    this.datasetHandler = new DatasetHandler(wkr);

  }

  /**
   * A preprocessor is a function which is given dataset context object, and
   * and the visualization settings given by an analyzer (or by person editing dashboard), and it generates the precise format object required by the given visualizations.
   * take a long time/be computationally intensive. It is permitted to output a visualization of a different type than it is given.
   *
   * @param {*} vistype The visualization type to handle
   * @param {*} f A function that performs preprocessing
   */
  addPreprocessor(vistype, f) {
    customPreprocessors.set(vistype, f);
  }

  /**
   * Visualization analyzers are functions that given a dataset context object, as well as the visualizations that have been computed so far,
   * decides which visualizations to use and how to set them up. As an example, given a numeric timeseries, an analyzer might
   * output the settings necessary to view the data as a line plot.
   *
   * @param {string} name The name of the visualization
   * @param {function} f The function that analyzes the data and sets up the visualization
   */
  addVisualization(name,f) {
    this.datasetHandler.addVisualization({name,f});
  }

  addAnalysisFunction(name,f) {
    addAnalysisFunction(name,f);
  }
  initDatasetContext(f) {
    initDatasetContext(f);
  }
}
export default TimeseriesInjector;
