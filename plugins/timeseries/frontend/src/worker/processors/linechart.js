import { LTTB } from "../../../dist/downsample.mjs";
// The colors supported by object views.
const objectColors = [
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

// Colors to display when showing a single time series
const singleColors = {
  low: "rgba(0,92,158,0.6)",
  high: "rgba(0,0,0,0.1)",
  background: "rgba(66,134,244,0.4)",
  border: "rgba(0,92,158,0.4)"
};

function computeDataset(
  d,
  f = d => d.d,
  name,
  color,
  showDuration,
  downsample
) {
  //console.log(d);
  let showLine = d.length < 500;
  let dataset = new Array(d.length);
  let isbool = true;
  let pointColor = d.length > 10000 ? color.high : color.low;

  if (showDuration) {
    isbool = false; // Datasets with duration are processed as non-boolean
    dataset = new Array(d.length * 3); // Start point, endpoint, and a null to break the line

    pointColor = new Array(d.length * 3); // point colors need to be set so that only start point is colored
    for (let i = 0; i < d.length; i++) {
      let data = f(d[i]);
      // We check if the dataset is boolean - in which case we draw a stepped line
      if (typeof data === "boolean") {
        if (data === false) {
          data = 0;
        } else {
          data = 1;
        }
      } else {
        isbool = false;
      }
      dataset[i * 3] = data;
      dataset[i * 3 + 1] = data;
      dataset[i * 3 + 2] = NaN;

      pointColor[i * 3] = color.low;
      pointColor[i * 3 + 1] = "transparent";
      pointColor[i * 3 + 2] = "transparent";
    }
  } else {
    for (let i = 0; i < d.length; i++) {
      let data = f(d[i]);
      // We check if the dataset is boolean - in which case we draw a stepped line
      if (typeof data === "boolean") {
        if (data === false) {
          data = 0;
        } else {
          data = 1;
        }
      } else {
        isbool = false;
      }
      if (downsample > 0) {
        // If downsampling, the dataset creates new time points
        dataset[i] = { x: d[i].t * 1000, y: data };
      } else {
        dataset[i] = data;
      }
    }
    if (downsample > 0) {
      dataset = LTTB(dataset, downsample);
    }
  }
  let shouldFill = isbool || showDuration ? true : d.length < 50;
  return {
    label: name,
    data: dataset,
    lineTension: 0,
    // For nicer displaying, we don't add a fill color when we have enough datapoints,
    // and when we have a lot of data, we turn into a scatter chart. For booleans, though,
    // we always use a fill, so that they are more visible.
    fill: shouldFill,
    showLine: isbool ? true : showLine,
    steppedLine: isbool,
    backgroundColor: shouldFill ? color.background : "transparent",
    borderColor: color.border,
    pointBackgroundColor: pointColor,
    pointBorderColor: pointColor,
    pointRadius: d.length > 500 ? (d.length > 10000 ? 1 : 2) : 3
  };
}

async function process(object, d) {
  if (d.length < 2) {
    return {};
  }
  if (!d.every(dp => !isNaN(dp.d))) {
    return {}; // This currently disallows objects
  }
  let legend = false;
  let datasetobj = {};
  let downsample = d.length > 50000 ? 50000 : 0;
  let showDuration = d.length < 500 && d.some(dp => dp.td > 0);
  // Prepare the labels
  let labels = new Array(d.length);
  if (showDuration) {
    labels = new Array(d.length * 3);
    for (let i = 0; i < d.length; i++) {
      labels[i * 3] = d[i].t * 1000;
      labels[i * 3 + 1] =
        (d[i].t + (d[i].td !== undefined ? d[i].td : 0)) * 1000;
      labels[i * 3 + 2] = labels[i * 3 + 1];
    }
  } else {
    for (let i = 0; i < d.length; i++) {
      labels[i] = d[i].t * 1000;
    }
  }

  // We can have two types of data: the first is a numeric stream, the second an object containing numeric data.
  // We will generate a different time series for each key of the object
  if (typeof d[0].d === "object") {
    legend = true; // Do display a legend

    let keys = Object.keys(d[0].d);
    let dataset = new Array(keys.length);
    for (let i = 0; i < keys.length; i++) {
      dataset[i] = computeDataset(
        d,
        p => p.d[keys[i]],
        keys[i],
        objectColors[i],
        showDuration,
        downsample
      );
    }
    datasetobj = { datasets: dataset };
  } else {
    datasetobj = {
      datasets: [
        computeDataset(d, p => p.d, "", singleColors, showDuration, downsample)
      ]
    };
  }
  if (!downsample) {
    datasetobj.labels = labels;
  }
  let bigchart = d.length > 10000;
  return {
    lineplot: {
      weight: 9,
      title: "Line Plot",
      view: "chartjs",
      data: {
        type: "line",
        options: {
          responsive: true,
          animation: {
            duration: 0 // general animation time
          },
          hover: {
            animationDuration: 0 // duration of animations when hovering an item
          },
          events: bigchart
            ? ["click"]
            : ["mousemove", "mouseout", "click", "touchstart", "touchmove"],
          responsiveAnimationDuration: 0, // animation duration after a resize
          legend: {
            display: legend
          },
          scales: {
            xAxes: [
              {
                type: "time",
                position: "bottom"
              }
            ]
          }
        },
        data: datasetobj
      }
    }
  };
}

export default process;
