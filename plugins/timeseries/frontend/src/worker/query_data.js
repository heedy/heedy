import { isNumeric, isBoolean } from "../analysis.mjs";

class QueryData {
  constructor(wkr, qobj, dataset) {
    this.qobj = qobj;
    this.worker = wkr;

    // The query that was run
    this.query = qobj.query;
    // The dataset that results from the query
    this.dataset = dataset;

    // Add on some basic analysis functions to the dataset, which will be shared
    // among all analyzers/preprocessors, so that they only need to be computed once, and their result is cached

    // internal cache for function results
    this._isNumeric = this.dataset.map((d) => null);
    this._isBoolean = this.dataset.map((d) => null);
    this._hasDuration = this.dataset.map((d) => null);

    for (let i = 0; i < this.dataset.length; i++) {
      this.dataset[i].isNumeric = () => {
        if (this._isNumeric[i] === null) {
          this._isNumeric[i] = isNumeric(this.dataset[i]);
        }
        return this._isNumeric[i];
      };
      this.dataset[i].isBoolean = () => {
        if (this._isBoolean[i] === null) {
          this._isBoolean[i] = isBoolean(this.dataset[i]);
        }
        return this._isBoolean[i];
      };
      this.dataset[i].hasDuration = () => {
        if (this._hasDuration[i] === null) {
          this._hasDuration[i] = this.dataset[i].some((dp) => dp.dt > 0);
        }
        return this._hasDuration[i];
      };
    }
  }

  /**
   * Gets the object specified by the ID. Remembers that the object was queried,
   * and has the query automatically re-run if the object changes.
   * @param {string} oid
   */
  getObject(oid) {
    // First check if we're subscribed to the object
  }
}

export default QueryData;
