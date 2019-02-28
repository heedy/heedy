/*
The CSVView displays the currently queried data as a CSV text, that can be copied to clipboard and imported
into excel and such
*/

import React, { Component } from "react";
import PropTypes from "prop-types";
import moment from "moment";

import DataUpdater from "../components/DataUpdater";

import "codemirror/lib/codemirror.css";
import "codemirror/theme/monokai.css";
import CodeMirror from "react-codemirror";

class CSVView extends DataUpdater {
  static propTypes = {
    data: PropTypes.arrayOf(PropTypes.object).isRequired,
    state: PropTypes.object.isRequired,
    setState: PropTypes.func.isRequired
  };

  // transformDataset is required for DataUpdater to set up the modified state data
  transformDataset(d) {
    let dataset = "";
    let dateFormat = "YYYY-MM-DD HH:mm:ss";

    if (d.length > 0) {
      // In order to show columns in the data table, we first check if the datapoints are objects...
      // If they are, then we generate the table so that the object is the columns
      if (d[0].d !== null && typeof d[0].d === "object") {
        dataset = "Timestamp";
        Object.keys(d[0].d).map(key => {
          dataset += "," + key.capitalizeFirstLetter();
        });
        dataset += "\n";

        for (let i = 0; i < d.length; i++) {
          dataset += moment.unix(d[i].t).format(dateFormat);
          Object.keys(d[i].d).map(key => {
            dataset += ", " + JSON.stringify(d[i].d[key], undefined, 2);
          });
          dataset += "\n";
        }
      } else {
        dataset = "Timestamp,Data\n";
        for (let i = 0; i < d.length; i++) {
          dataset +=
            moment.unix(d[i].t).format(dateFormat) +
            "," +
            JSON.stringify(d[i].d) +
            "\n";
        }
      }
    }

    return dataset;
  }

  render() {
    return (
      <CodeMirror
        value={this.data}
        options={{
          lineWrapping: true,
          readOnly: true,
          mode: "text/plain"
        }}
      />
    );
  }
}
export default CSVView;
