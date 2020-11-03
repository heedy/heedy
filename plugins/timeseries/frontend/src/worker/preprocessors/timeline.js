
import query from "../../analysis.mjs";
import moment from "../../../dist/moment.mjs";

const day = 60 * 60 * 24;

const datesAreOnSameDay = (first, second) =>
    first.getFullYear() === second.getFullYear() &&
    first.getMonth() === second.getMonth() &&
    first.getDate() === second.getDate();

const endDate = (d) =>
    new Date(d.getFullYear(), d.getMonth(), d.getDate(), 23, 59, 59, 999);
const nextDate = (d) =>
    new Date(d.getFullYear(), d.getMonth(), d.getDate() + 1, 0, 0, 0, 0);

function perDay(ts) {
    let days = [];
    let curday = [];
    let i = 0;
    let dp = ts[i];
    let curDate = new Date(dp.t * 1000);
    while (true) {
        let startTime = new Date(dp.t * 1000);
        let endTime = new Date(1000 * (dp.t + (dp.dt === undefined ? 0 : dp.dt)));

        if (datesAreOnSameDay(curDate, startTime)) {
            // We add the datapoint to the current day array
            if (datesAreOnSameDay(curDate, endTime)) {
                // The entire datapoint fits in the same day, so just add it whole
                curday.push(dp);
                i += 1;
                if (i >= ts.length) {
                    days.push({ date: curDate, data: curday });
                    break;
                }
                dp = ts[i];
            } else {
                // The datapoint does NOT fit in the day, so split it into a portion in the day, and a portion outside
                let dt = nextDate(curDate).getTime() / 1000 - dp.t;
                curday.push({
                    t: dp.t,
                    td: dt,
                    d: dp.d,
                });
                dp = {
                    t: nextDate(curDate).getTime() / 1000,
                    td: dp.dt - dt,
                    d: dp.d,
                };
            }
        } else {
            // The next day!
            days.push({ date: curDate, data: curday });
            curday = [];
            curDate = nextDate(curDate);
        }
    }
    return days;
}

function prepareTimeline(extractor, dp) {
    return {
        timeRange: [
            new Date(dp.t * 1000),
            new Date(1000 * (dp.t + (dp.dt === undefined ? 0 : dp.dt)))
        ],
        val: extractor(dp)
    };
}

function resetYear(d) {
    d.setFullYear(1970, 1, 2);
    return d;
}

function preprocess(qd, visualization) {
    let data = null;

    if (visualization.config.scale == "multi") {
        data = [{
            group: "",
            data: visualization.config.data.map(d => {
                let extractor = query(d.q);
                return {
                    label: d.label,
                    data: qd.dataset[d.series].map(dp => prepareTimeline(extractor, dp))
                }
            })
        }];
    } else if (visualization.config.data.length == 1) {
        // The data is per-day. If there is just 1 series, show the times on right hand labels
        let extractor = query(visualization.config.data[0].q);
        data = [{
            group: "",
            data: perDay(qd.dataset[visualization.config.data[0].series]).map(dval => ({
                label: moment(dval.date).format("YYYY-MM-DD"),
                data: dval.data.map(dp => prepareTimeline(extractor, dp)).map(dp => {
                    resetYear(dp.timeRange[0]);
                    resetYear(dp.timeRange[1]);
                    return dp;
                })
            }))
        }];
    } else {
        // There are multiple series, so we group them by date

        let groups = {};

        let sperday = visualization.config.data.map(s => {
            let dateGroup = {};
            let extractor = query(s.q);
            perDay(qd.dataset[s.series]).forEach(dval => {
                let curday = moment(dval.date).format("YYYY-MM-DD");
                groups[curday] = true;
                dateGroup[curday] = {
                    label: s.label,
                    data: dval.data.map(dp => prepareTimeline(extractor, dp)).map(dp => {
                        resetYear(dp.timeRange[0]);
                        resetYear(dp.timeRange[1]);
                        return dp;
                    })
                }
            });
            return dateGroup;
        });

        let dates = Object.keys(groups);
        dates.sort()
        data = dates.map(k => ({
            group: k,
            data: sperday.map((s, i) => {
                if (s[k] === undefined) {
                    return {
                        label: visualization.config.data[i].label,
                        data: []
                    };
                }
                return s[k];
            })
        }));
    }

    return {
        ...visualization,
        config: {
            ...visualization.config,
            data: data
        }
    };
}

export default preprocess;