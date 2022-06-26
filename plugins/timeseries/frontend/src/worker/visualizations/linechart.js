// The colors supported by object views.
const multiSeriesColors = [{
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
    low: "rgba(0,92,158,0.6)", // Low number of datapoints
    high: "rgba(0,0,0,0.1)", // high number of datapoints
    background: "rgba(66,134,244,0.4)",
    border: "rgba(0,92,158,0.4)",
};


const chartjsSettings = (isLarge, aspectRatio, ylabel, ykey, datasets,syncT=false) => ({
    type: "line",
    options: {
        responsive: true,
        aspectRatio: aspectRatio,
        animation: {
            duration: 0, // general animation time
        },
        parsing: {
            xAxisKey: "t",
            yAxisKey: ykey,
        },
        hover: {
            animationDuration: 0, // duration of animations when hovering an item
        },
        events: isLarge ? ["click"] : ["mousemove", "mouseout", "click", "touchstart", "touchmove"],
        responsiveAnimationDuration: 0, // animation duration after a resize
        plugins: {
            legend: {
                display: false,
            }
        },
        scales: {
            x: {
                type: "time",
                position: "bottom",
                ...(syncT?{min:"${{data.minTimestamp()}}",max:"${{data.maxTimestamp()}}"}:{}),
            },
            y: {
                type: "linear",
                position: "left",
                title: {
                    display: ylabel != "",
                    text: ylabel,
                },
            },
        },
    },
    data: {
        datasets: datasets,
    },
});

/**
 * 
 * @param {*} c 
 * @param {*} da 
 * @param {*} d 
 * @param {*} colors 
 * @param {*} key 
 * @param {*} yid 
 * @param {*} label 
 * @returns 
 */
function generateDataset(c, da, d, colors, label) {
    const len = d.nonNull();
    const isbool = d.type() == "boolean";
    const showDuration = len < 500 && da.dt.nonNull() > 0;
    const fillBackground = isbool || showDuration || len < 50;
    let pointColor = len > 5000 ? colors.high : colors.low;

    const dataPath = c.tpls('d', ...d.path);
    const dex = `filterNull(data[${d.index}],${dataPath})`;
    let dataExtractor = dex;
    if (len > 50000) {
        dataExtractor = `downsample(${dataExtractor},50000,${dataPath})`;
    } else if (showDuration) {
        // If we show the duration, we want to display the points with explicit duration,
        // so the data has to be expanded with explicit points for start/end of range,
        // and we need to set the colors of each point separately.
        dataExtractor = `explicitDuration(${dataExtractor},{'separator':null,'offset':0.001})`
        pointColor = c.tpl(`[...${dex}].fill([${c.tpls(pointColor)},'transparent','transparent']).flat()`);
    }
    const out = {
        lineTension: 0,
        label: label,
        showLine: isbool || len < 500,
        steppedLine: isbool,
        pointRadius: len > 500 ? (len > 10000 ? 1 : 2) : 3,
        fill: fillBackground,
        backgroundColor: fillBackground ? colors.background : "transparent",
        borderColor: colors.border,
        pointBackgroundColor: pointColor,
        pointBorderColor: pointColor,
        data: c.tpl(dataExtractor)
    };

    return out;
}

function linechart(c, vis) {
    if (c.data.length == 0 || c.data.length > 4 || !c.data.every(da => da.length > 2)) {
        return;
    }

    let config = null;
    if (!c.data.every(da => da.d.type() == "number")) {
        // The data is not numeric. We can also handle a single series of objects
        if (c.data.length != 1 || c.data[0].d.type() != "object") {
            return;
        }
        const da = c.data[0];
        const k = da.d.keys();
        const usefulKeys = Object.keys(k)
            .filter(kv => k[kv] >= da.length / 2)
            .filter(kv => {
                const t = da.d(kv).type();
                return t == "number" || t == "boolean";
            });

        // if there is nothing to display or we're looking at coordinates, don't display anything
        if (usefulKeys.length == 0 || k["latitude"] !== undefined || k["longitude"] !== undefined) {
            return;
        }

        // Sort by key name, giving the same color for iterated series
        usefulKeys.sort();

        if (usefulKeys.length > 4) {
            // If there are many keys, only show the 4 most useful
            usefulKeys.sort((a, b) => k[b] - k[a]);
            usefulKeys = usefulKeys.slice(0, 4);
        }

        // Finally, construct the plots, one for each key of the object

        config = usefulKeys.map((kk, i) => chartjsSettings(k[kk] > 5000, usefulKeys.length, kk, `d.${kk}`, [
            generateDataset(c, da, da.d(kk), multiSeriesColors[i], kk)
        ]));


    } else if (c.data.length == 1) {
        const da = c.data[0];
        config = [chartjsSettings(da.length > 5000, 1.2, "", "d", [generateDataset(c, da, da.d, singleSeriesColor, "")])]
    } else {
        config = c.data.map((da,i) => {
            const label = c.getSeriesLabelTemplate(i);
            return chartjsSettings(da.length > 5000, c.data.length,label,"d",[generateDataset(c, da, da.d, multiSeriesColors[i], label)],true)
        });
    }

    vis.set("linechart", {
        weight: 9,
        title: "Raw Plot",
        type: "chartjs",
        config: config
    });
}

export default linechart;