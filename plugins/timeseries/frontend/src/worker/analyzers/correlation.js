// The colors supported by object views. They are shifted by 1 with respect to linechart colors,
// to reflect the correlation being used
const multiSeriesColors = [
  {
    low: "rgba(255, 99, 132,0.6)",
    high: "rgba(255, 99, 132,0.1)",
    background: "rgba(255, 99, 132,0.1)",
    border: "rgba(255, 99, 132,0.4)",
  },
  {
    low: "rgba(54, 162, 235,0.6)",
    high: "rgba(54, 162, 235,0.1)",
    background: "rgba(54, 162, 235,0.1)",
    border: "rgba(54, 162, 235,0.4)",
  },
  {
    low: "rgba(255, 206, 86,0.6)",
    high: "rgba(255, 206, 86,0.1)",
    background: "rgba(255, 206, 86,0.1)",
    border: "rgba(255, 206, 86,0.4)",
  },
  {
    low: "rgba(75, 192, 192,0.6)",
    high: "rgba(75, 192, 192,0.1)",
    background: "rgba(75, 192, 192,0.1)",
    border: "rgba(75, 192, 192,0.4)",
  },
];

const singleSeriesColor = {
  low: "rgba(0,92,158,0.6)",
  high: "rgba(0,0,0,0.1)",
  background: "rgba(66,134,244,0.4)",
  border: "rgba(0,92,158,0.4)",
};

let chartjsSettings = (isLarge, aspectRatio, xlabel, ylabel, datasets) => ({
  type: "line",
  options: {
    responsive: true,
    aspectRatio: aspectRatio,
    animation: {
      duration: 0, // general animation time
    },
    hover: {
      animationDuration: 0, // duration of animations when hovering an item
    },
    events: isLarge
      ? ["click"]
      : ["mousemove", "mouseout", "click", "touchstart", "touchmove"],
    responsiveAnimationDuration: 0, // animation duration after a resize
    legend: {
      display: false,
    },
    scales: {
      xAxes: [
        {
          type: "linear",
          position: "bottom",
          scaleLabel: {
            display: xlabel != "",
            labelString: xlabel,
          },
        },
      ],
      yAxes: [
        {
          type: "linear",
          id: "y0",
          position: "left",
          scaleLabel: {
            display: ylabel != "",
            labelString: ylabel,
          },
        },
      ],
    },
  },
  data: {
    datasets: datasets,
  },
});

function generateDataset(d, x, y, colors, idx) {
  // Generate a chartjs config for this specific data array
  let pointColor = d.length > 5000 ? colors.high : colors.low;

  return {
    lineTension: 0,
    label: `Series ${idx + 1}`,
    showLine: false,
    pointRadius: d.length > 500 ? (d.length > 10000 ? 1 : 2) : 3,
    fill: false,
    backgroundColor: "transparent",
    borderColor: colors.border,
    pointBackgroundColor: pointColor,
    pointBorderColor: pointColor,
    data: {
      // The data object is replaced with query data
      series: idx,
      x: x,
      y: y,
      downsample: d.length > 50000 ? 50000 : 0,
      withDuration: false,
      removeNull: true,
    },
  };
}

function analyze(qd) {
  if (
    qd.dataset.length != 1 ||
    qd.dataset[0].dataType() != "object" ||
    qd.dataset[0].length <= 1
  ) {
    return {}; // We only handle objects for correlation scatterplots
  }

  let d = qd.dataset[0];
  let k = d.keys();
  // Filter out the keys with less than half data, and which are not numbers
  let usefulKeys = Object.keys(k)
    .filter((kv) => k[kv] >= d.length / 2)
    .filter((kv) => d.keyType(kv) === "number"); // Only accept numbers, not booleans

  usefulKeys.sort(); // Sort alphabetically by key

  if (
    usefulKeys.length < 2 ||
    usefulKeys.length > 5 ||
    k["latitude"] !== undefined ||
    k["longitude"] !== undefined
  ) {
    return {};
  }

  let xkey = usefulKeys[0];
  let yKeys = usefulKeys.splice(1, usefulKeys.length);

  let charts = [];
  if (yKeys.length > 1) {
    charts = yKeys.map((yk, i) =>
      chartjsSettings(k[xkey] > 5000, yKeys.length, xkey, yk, [
        generateDataset(d, ["d", xkey], ["d", yk], multiSeriesColors[i], 0),
      ])
    );
  } else {
    charts = [
      chartjsSettings(k[xkey] > 5000, 1, xkey, yKeys[0], [
        generateDataset(d, ["d", xkey], ["d", yKeys[0]], singleSeriesColor, 0),
      ]),
    ];
  }

  return {
    correlation: {
      weight: 10,
      title: "Correlation",
      visualization: "chartjs",
      config: {
        charts: charts,
      },
    },
  };
}

export default analyze;
