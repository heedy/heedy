
import query, { dq, dtq } from "../../analysis.mjs";

// The colors supported by object views.
const multiSeriesColors = [
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
  {
    low: "rgba(255, 206, 86,0.6)",
    high: "rgba(255, 206, 86,0.1)",
    background: "rgba(255, 206, 86,0.1)",
    border: "rgba(255, 206, 86,0.4)",
  },
];

const singleSeriesColor = {
  low: "rgba(0,92,158,0.6)",
  high: "rgba(0,0,0,0.1)",
  background: "rgba(66,134,244,0.4)",
  border: "rgba(0,92,158,0.4)",
};

let chartjsSettings = (isLarge, aspectRatio, ylabel, datasets) => ({
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

function generateDataset(d, y, colors, idx, yid) {
  // Generate a chartjs config for this specific data array
  let isbool = dq.isBoolean(d);
  let showDuration = d.length < 500 && dtq.sum(d) > 0;
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
    yAxisID: yid,
    data: {
      // The data object is replaced with query data
      series: idx,
      x: ["t"],
      y: y,
      downsample: d.length > 50000 ? 50000 : 0,
      withDuration: showDuration,
      removeNull: true,
    },
  };
}

function analyze(qd) {
  if (qd.dataset.length == 0 || qd.dataset.length > 4 || !qd.dataset.every(da => da.length > 2)) {
    return {};
  }

  let charts = null;

  if (!qd.dataset.every((da) => dq.isNumeric(da))) {
    // The data is not numeric. Find keys if it is an object
    if (qd.dataset.length != 1 || dq.dataType(qd.dataset[0]) != "object") {
      return {}; // We only handle objects for single series
    }

    let d = qd.dataset[0];
    let k = dq.keys(d);
    // Filter out the keys with less than half data, and which are not numbers
    let usefulKeys = Object.keys(k)
      .filter((kv) => k[kv] >= d.length / 2)
      .filter((kv) => {
        let kt = query(["d", kv]).dataType(d);
        return kt === "number" || kt === "boolean";
      });

    // Sort by key name - this gives same color to correlation series as raw series
    usefulKeys.sort();

    if (
      usefulKeys.length == 0 ||
      k["latitude"] !== undefined ||
      k["longitude"] !== undefined
    ) {
      return {};
    }

    if (usefulKeys.length > 4) {
      // Sort by number of datapoints
      usefulKeys.sort((a, b) => k[b] - k[a]);
      usefulKeys = usefulKeys.slice(0, 4);
    }

    // OK, so now construct the plots using only the useful keys
    charts = usefulKeys.map((kv, i) =>
      chartjsSettings(k[kv] > 5000, usefulKeys.length, kv, [
        generateDataset(d, ["d", kv], multiSeriesColors[i], 0, "y0"),
      ])
    );
  } else {
    if (qd.dataset.length == 1) {
      charts = [
        chartjsSettings(qd.dataset[0].length > 5000, 1.2, "", [
          generateDataset(qd.dataset[0], ["d"], singleSeriesColor, 0, "y0"),
        ]),
      ];
    } /*else if (qd.dataset.length == 2) {
      charts = [
        chartjsSettings(
          qd.dataset[0].length > 5000 || qd.dataset[1].length > 5000,
          1.2,
          "Series 1",
          [
            generateDataset(
              qd.dataset[0],
              ["d"],
              multiSeriesColors[0],
              0,
              "y0"
            ),
            generateDataset(
              qd.dataset[1],
              ["d"],
              multiSeriesColors[1],
              1,
              "y1"
            ),
          ]
        ),
      ];

      charts[0].options.legend.display = true;
      charts[0].options.scales.yAxes.push({
        type: "linear",
        id: "y1",
        position: "right",
        scaleLabel: {
          display: true,
          labelString: "Series 2",
        },

        gridLines: {
          drawOnChartArea: false,
        },
      });
    }*/ else {
      charts = qd.dataset.map((d, i) =>
        chartjsSettings(d.length > 5000, qd.dataset.length, `Series ${i + 1}`, [
          generateDataset(d, ["d"], multiSeriesColors[i], i, "y0"),
        ])
      );
    }
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
