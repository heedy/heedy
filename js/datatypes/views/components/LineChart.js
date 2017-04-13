/*
This shows a line chart of the data given
*/

import React, { PropTypes } from "react";
import DataTransformUpdater from "./DataUpdater";

import { Line } from "react-chartjs-2";
import moment from "moment";

export class LineChart extends DataTransformUpdater {
  static propTypes = {
    data: PropTypes.arrayOf(PropTypes.object).isRequired,
    transform: PropTypes.string
  };

  // The colors supported by object views.
  static objectColors = [
    {
      low: "rgba(54, 162, 235,0.6)",
      high: "rgba(54, 162, 235,0.1)",
      background: "rgba(54, 162, 235,0.1)",
      border: "rgba(54, 162, 235,0.4)"
    },
    {
      low: "rgba(255, 206, 86,0.6)",
      high: "rgba(255, 206, 86,0.1)",
      background: "rgba(255, 206, 86,0.1)",
      border: "rgba(255, 206, 86,0.4)"
    },
    {
      low: "rgba(75, 192, 192,0.6)",
      high: "rgba(75, 192, 192,0.1)",
      background: "rgba(75, 192, 192,0.1)",
      border: "rgba(75, 192, 192,0.4)"
    },
    {
      low: "rgba(255, 99, 132,0.6)",
      high: "rgba(255, 99, 132,0.1)",
      background: "rgba(255, 99, 132,0.1)",
      border: "rgba(255, 99, 132,0.4)"
    }
  ];

  static singleColors = {
    low: "rgba(0,92,158,0.6)",
    high: "rgba(0,0,0,0.1)",
    background: "rgba(66,134,244,0.4)",
    border: "rgba(0,92,158,0.4)"
  };

  // transformDataset is required for DataUpdater to set up the modified state data
  transformDataset(d) {
    // We can have two types of data: the first is a numeric stream, the second an object containing numeric data.
    // We will generate a different time series for each key of the object
    if (d.length == 0) {
      return {
        datasets: [
          {
            data: []
          }
        ]
      };
    }

    if (typeof d[0].d === "object") {
      this.obj = true; // Do display a legend

      let keys = Object.keys(d[0].d);
      let dataset = new Array(keys.length);
      for (let i = 0; i < keys.length; i++) {
        dataset[i] = this.computeDataset(
          d,
          p => p.d[keys[i]],
          keys[i],
          LineChart.objectColors[i]
        );
      }
      return { datasets: dataset };
    }

    this.obj = false; // Don't display the legend
    return {
      datasets: [this.computeDataset(d, p => p.d, "", LineChart.singleColors)]
    };
  }

  /**
     * 
     * @param {*} d The datapoint array to process
     * @param {*} f A function that transforms a single datapoint's data into the wanted format
     * @param {*} name The name to show in the legend
     * @param {*} color The color to use for displaying this series
     */
  computeDataset(d, f = d => d.d, name, color) {
    let dataset = new Array(d.length);

    // We check if the dataset is boolean - in which case we draw a stepped line
    let isbool = true;

    for (let i = 0; i < d.length; i++) {
      let data = f(d[i]);
      if (typeof data === "boolean") {
        if (data === false) {
          data = 0;
        } else {
          data = 1;
        }
      } else {
        isbool = false;
      }
      dataset[i] = {
        x: moment.unix(d[i].t),
        y: data
      };
    }

    let pointColor = d.length > 500 ? color.high : color.low;

    return {
      label: name,
      data: dataset,
      lineTension: 0,
      // For nicer displaying, we don't add a fill color when we have enough datapoints,
      // and when we have a lot of data, we turn into a scatter chart. For booleans, though,
      // we always use a fill, so that they are more visible.
      fill: isbool ? true : d.length < 50,
      showLine: isbool ? true : d.length < 500,
      steppedLine: isbool,
      backgroundColor: color.background,
      borderColor: color.border,
      pointBackgroundColor: pointColor,
      pointBorderColor: pointColor,
      pointRadius: d.length > 500 ? 2 : 3
    };
  }

  render() {
    return (
      <Line
        data={this.data}
        options={{
          legend: {
            display: this.obj
          },
          scales: {
            xAxes: [
              {
                type: "time",
                position: "bottom"
              }
            ]
          },
          animation: false
        }}
      />
    );
  }
}

export default LineChart;

// generate creates a new view that displays a line chart. The view object is set up
// so that it is totally ready to be passed as a result of the shower function
export function generateLineChart(transform = "") {
  let component = LineChart;

  // If we're given a transform, wrap the LineChart so that we can pass transform into the class.
  if (transform !== "") {
    component = React.createClass({
      render: function() {
        return <LineChart {...this.props} transform={transform} />;
      }
    });
  }

  return { initialState: {}, component: component, width: "expandable-half" };
}
