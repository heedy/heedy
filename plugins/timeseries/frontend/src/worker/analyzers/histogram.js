import query, { dq } from "../../analysis.mjs";


let chartOptions = (colors, bins, series, q, label, aspectRatio) => ({
    type: "bar",
    options: {
        responsive: true,
        aspectRatio: aspectRatio,
        animation: {
            duration: 0
        },
        hover: {
            animationDuration: 0, // duration of animations when hovering an item
        },
        events: ["mousemove", "mouseout", "click", "touchstart", "touchmove"],
        responsiveAnimationDuration: 0, // animation duration after a resize
        legend: {
            display: false,
        },
        scales: {
            yAxes: [
                {
                    type: "linear",
                    id: "y0",
                    position: "left",
                    scaleLabel: {
                        display: label != "",
                        labelString: label,
                    },
                },
            ],
        }
    },
    data: {
        labels: {
            bins: bins
        },
        datasets: [{
            backgroundColor: colors.background,
            borderWidth: 1,
            borderColor: colors.border,
            data: {
                key: series,
                x: q
            }
        }]
    }
})

// The colors supported by object views.
const multiSeriesColors = [
    {
        background: "rgba(75, 192, 192,0.5)",
        border: "rgba(75, 192, 192,0.6)",
    },
    {
        background: "rgba(255, 99, 132,0.5)",
        border: "rgba(255, 99, 132,0.6)",
    },
    {
        background: "rgba(54, 162, 235,0.5)",
        border: "rgba(54, 162, 235,0.6)",
    },
    {
        background: "rgba(255, 206, 86,0.5)",
        border: "rgba(255, 206, 86,0.6)",
    },
];


const singleSeriesColor = {
    background: "rgba(66,134,244,0.6)",
    border: "rgba(0,92,158,0.7)",
};

function getNumericHist(qd, series, q, colors) {
    // Given a series, return the appropriate chart object
    let bins = 20;

    return chartOptions(colors, bins, series, q, true);

}

function prepareObjectHist(dp) {
    // Given a single datapoint that contains an object, where each key is associated with a number,
    // draw the histogram
}



function analyze(qd) {
    if (qd.keys.length == 0 || qd.keys.length > 4) {
        return {};
    }

    let charts = null;

    // If the dataset is just an object
    if (qd.dataset_array.length == 1 && qd.dataset_array[0].length > 40 && dq.dataType(qd.dataset_array[0]) == "object") {
        let d = qd.dataset_array[0];
        let k = dq.keys(d);
        // Filter out the keys with less than half data, and which are not numbers
        let usefulKeys = Object.keys(k)
            .filter((kv) => k[kv] >= d.length / 2)
            .filter((kv) => (query(["d", kv]).dataType(d) == "number"));

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
        charts = usefulKeys.map((kv, i) => chartOptions(multiSeriesColors[i], k[kv] > 500 ? 20 : 10, qd.keys[0], ["d", kv], kv, usefulKeys.length));
    } else if (qd.dataset_array.every(ds => ds.length > 40 && dq.isNumeric(ds))) {
        // The dataset is a histogram for each
        charts = qd.dataset_array.map((ds, i) => chartOptions(qd.dataset_array.length == 1 ? singleSeriesColor : multiSeriesColors[i], ds.length > 500 ? 20 : 10, qd.keys[i], ["d"], qd.dataset_array.length == 1 ? "Number of Datapoints" : qd.keys[i], qd.dataset_array.length == 1 ? 1.2 : qd.dataset_array.length))

    } else {
        return {};
    }

    return {
        histogram: {
            weight: 11,
            title: "Histogram",
            visualization: "histogram",
            config: {
                charts: charts
            }
        }
    }
}

export default analyze;