import query from "../../analysis.mjs";



function printNum(num, decimals) {
    return num.toFixed(decimals).replace(/\.?0*$/, "");
}

function processChart(qd, chart) {
    // Get the bounds range of the data
    let tmin = Infinity;
    let tmax = -Infinity;
    chart.data.datasets.forEach((dss) => {

        let ds = dss.data;
        let q = query(ds.x);
        let d = qd.dataset[ds.key];
        let m = q.min(d);
        if (m < tmin) {
            tmin = m;
        }
        m = q.max(d);
        if (m > tmax) {
            tmax = m;
        }
    });

    // Get the labels
    let bins = chart.data.labels.bins;

    // Make bin widths cleaner
    let binwidth = (tmax - tmin) / bins;
    if (binwidth > 80) {
        tmin = Math.floor(tmin / 100) * 100;
        tmax = Math.ceil(tmax / 100) * 100;
    } else if (binwidth > 8) {
        tmin = Math.floor(tmin / 10) * 10;
        tmax = Math.ceil(tmax / 10) * 10;
    } else if (binwidth > 0.8) {
        tmin = Math.floor(tmin);
        tmax = Math.ceil(tmax);
    }
    binwidth = (tmax - tmin) / bins;

    let labels = new Array(bins);
    for (let i = 0; i < bins; i++) {
        labels[i] = `[${printNum(tmin + i * binwidth, 2)},${printNum(tmin + (i + 1) * binwidth, 2)})`;
    }



    return {
        ...chart,
        data: {
            ...chart.data,
            labels: labels,
            datasets: chart.data.datasets.map(dss => {
                let ds = dss.data
                let q = query(ds.x);
                let hist = new Array(bins).fill(0);
                qd.dataset[ds.key].forEach(dp => {
                    let d = q(dp);
                    if (d != null) {
                        let binindex = Math.floor((d - tmin) / binwidth);
                        if (binindex >= 0 && binindex < bins) {
                            hist[binindex]++;
                        }
                    }
                });
                return {
                    ...dss,

                    data: hist
                }
            })
        }
    }
}

function preprocess(qd, visualization) {
    return {
        ...visualization,
        visualization: "chartjs",
        config: {
            ...visualization.config,
            charts: visualization.config.charts.map(c => processChart(qd, c))
        }
    };
}

export default preprocess;