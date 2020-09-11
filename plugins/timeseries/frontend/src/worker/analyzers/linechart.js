// The colors supported by object views.
const multiSeriesColors = [
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
];

const singleSeriesColor = {
  low: "rgba(0,92,158,0.6)",
  high: "rgba(0,0,0,0.1)",
  background: "rgba(66,134,244,0.4)",
  border: "rgba(0,92,158,0.4)",
};

let chartjsSettings = (isLarge, aspectRatio, datasets) => ({
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
          type: "time",
          position: "bottom",
        },
      ],
      yAxes: [
        {
          type: "linear",
          id: "y0",
          position: "left",
        },
      ],
    },
  },
  data: {
    datasets: datasets,
  },
});

function generateDataset(d, colors, idx, yid) {
  // Generate a chartjs config for this specific data array
  let isbool = d.isBoolean();
  let showDuration = d.length < 500 && d.hasDuration();
  let fillBackground = isbool || showDuration || d.length < 50;
  let pointColor = d.length > 5000 ? colors.high : colors.low;

  return {
    lineTension: 0,
    label: `Series ${idx + 1}`,
    showLine: isbool || d.length < 500,
    steppedLine: isbool,
    pointRadius: d.length > 500 ? (d.length > 10000 ? 1 : 2) : 3,
    fill: fillBackground,
    backgroundColor: fillBackground ? colors.background : "transparent",
    borderColor: colors.border,
    pointBackgroundColor: pointColor,
    pointBorderColor: pointColor,
    data: {
      // The data object is replaced with query data
      series: idx,
      x: ["t"],
      y: ["d"],
      downsample: d.length > 50000 ? 50000 : 0,
      withDuration: showDuration,
    },
  };
}

function analyze(qd) {
  if (
    (!qd.dataset.every((da) => da.isNumeric()) && qd.dataset.length == 0) ||
    qd.dataset.length > 4
  ) {
    return {};
  }

  let charts = null;

  if (qd.dataset.length == 1) {
    charts = [
      chartjsSettings(qd.dataset[0].length > 5000, 1, [
        generateDataset(qd.dataset[0], singleSeriesColor, 0, "y0"),
      ]),
    ];
  } else if (qd.dataset.length == 2) {
    charts = [
      chartjsSettings(
        qd.dataset[0].length > 5000 || qd.dataset[1].length > 5000,
        1,
        [
          generateDataset(qd.dataset[0], multiSeriesColors[0], 0, "y0"),
          generateDataset(qd.dataset[1], multiSeriesColors[1], 1, "y1"),
        ]
      ),
    ];

    charts[0].options.legend.display = true;
    charts[0].options.scales.yAxes.push({
      type: "linear",
      id: "y1",
      position: "right",
      gridLines: {
        drawOnChartArea: false,
      },
    });
  } else {
    charts = qd.dataset.map((d, i) =>
      chartjsSettings(d.length > 5000, qd.dataset.length, [
        generateDataset(d, multiSeriesColors[i], i, "y0"),
      ])
    );
  }

  return {
    linechart: {
      weight: 9,
      title: "Raw Plot",
      visualization: "chartjs",
      config: {
        charts: charts,
        syncX: true,
      },
    },
  };
}

export default analyze;
