import { getType, day, perDay, explicitDuration } from "../../analysis.mjs";
import moment from "../../../dist/moment.mjs";
import { LTTB } from "../../../dist/downsample.mjs";

function prepareTimeline(dp) {
  return {
    timeRange: [
      new Date(dp.t * 1000),
      new Date(1000 * (dp.t + (dp.dt === undefined ? 0 : dp.dt)))
    ],
    val: dp.d
  };
}

function resetYear(d) {
  d.setFullYear(1970, 1, 2);
  return d;
}

function process(o, ts) {
  if (ts.length < 10) {
    return {};
  }
  let dt = ts[ts.length - 1].t - ts[0].t;
  let dataType = getType(ts);
  if (dataType == "categorical") {
    if (dt <= 1.5 * day) {
      return {
        dayview: {
          weight: 10,
          title: "Timeline",
          view: "timeline",
          data: {
            discrete: true,
            leftMargin: 0,
            rightMargin: 0,
            timeFormat: "%Y-%m-%d %-I:%M:%S %p",
            scale: "multi",
            data: [
              {
                group: "",
                data: [
                  {
                    label: "",
                    data: ts.map(prepareTimeline)
                  }
                ]
              }
            ]
          }
        }
      };
    }

    // Otherwise, we show a per-day view
    let datas = perDay(ts).map(dval => ({
      label: moment(dval.date).format("YYYY-MM-DD"),
      data: dval.data.map(prepareTimeline).map(dp => {
        resetYear(dp.timeRange[0]);
        resetYear(dp.timeRange[1]);
        return dp;
      })
    }));

    return {
      dayview: {
        weight: 10,
        title: "Timeline",
        view: "timeline",
        data: {
          discrete: true,
          leftMargin: 0,
          rightMargin: 80,
          timeFormat: "%-I:%M:%S %p",
          scale: "day",
          data: [
            {
              group: "",
              data: datas
            }
          ]
        }
      }
    };
  }
  if (dataType == "number") {
    if (dt <= 2 * day || (dt / day) * 10 >= ts.length) {
      return {};
    }
    let datas = perDay(ts).map(dval => {
      let series = moment(dval.date).format("YYYY-MM-DD");
      let dur = explicitDuration(dval.data);
      if (dur.length > 1200) {
        // If more than a thousand points per day, downsample the timeseries

        dur = LTTB(
          dur.map(dp => ({ x: dp.t, y: dp.d })),
          dt > 20 * day ? (dt > 60 ? 200 : 500) : 1000
        ).map(dp => ({ t: dp.x, d: dp.y }));
      }
      return dur.map(dp => ({
        series: series,
        ts: resetYear(new Date(dp.t * 1000)),
        val: dp.d
      }));
    });

    let bands = 1;
    if (datas.length > 10) {
      bands = 4;
    }

    datas = datas.flat();

    // There's 2 days or more of data that is all numeric. Let's show a per-day view of the time series
    return {
      dayview: {
        weight: 10,
        title: "Per Day",
        view: "horizon",
        data: {
          label: bands <= 2,
          bands: bands,
          data: datas
        }
      }
    };
  }

  return {};
}

export default process;
