import { dq, dtq } from "../../analysis.mjs";

const day = 60 * 60 * 24;

async function analyze(qd) {
    if (qd.dataset_array.length == 0 || !qd.dataset_array.every(ds => dq.dataType(ds) == "categorical")) {
        return {};
    }
    // All the datasets are categorical, so let's display them!
    if (qd.dataset_array.every(ds => dtq.sum(ds) > 0)) {
        // They all have durations, so can be displayed on the swimlane timeline chart
        let scale = "day";
        let dt = 0;
        qd.dataset_array.forEach(ts => {
            let cdt = ts[ts.length - 1].t - ts[0].t;
            if (cdt > dt) {
                dt = cdt;
            }
        });
        if (dt < 1.5 * day) {
            scale = "multi";
        }

        return {
            timeline: {
                weight: 10,
                title: "Per-Day Timeline",
                visualization: "timeline",
                config: {
                    leftMargin: 0,
                    rightMargin: 0,
                    timeFormat: "%-I:%M:%S %p",
                    scale: scale,
                    data: qd.dataset_array.map((ds, i) => ({ key: qd.keys[i], q: ["d"], label: qd.dataset_array.length == 1 ? "" : qd.keys[i] }))
                }
            }
        }
    }

    return {};
}


export default analyze;