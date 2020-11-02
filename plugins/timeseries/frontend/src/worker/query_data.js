import { dq } from "../analysis.mjs";

class QueryData {
  constructor(wkr, qobj, dataset) {
    this.qobj = qobj;
    this.worker = wkr;

    // The query that was run
    this.query = qobj.query;
    // The dataset that results from the query
    this.dataset = dataset;
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
