import moment from "../../../dist/moment.mjs";

const day = 60 * 60 * 24;

const datesAreOnSameDay = (first, second) =>
  first.getFullYear() === second.getFullYear() &&
  first.getMonth() === second.getMonth() &&
  first.getDate() === second.getDate();

const endDate = d =>
  new Date(d.getFullYear(), d.getMonth(), d.getDate(), 23, 59, 59, 999);
const nextDate = d =>
  new Date(d.getFullYear(), d.getMonth(), d.getDate() + 1, 0, 0, 0, 0);

async function process(object, d) {
  if (d.length < 2) {
    return {};
  }
  // First, find out how many rows the plot needs
  let dt = d[d.length - 1].t - d[0].t;

  let labelSet = {};
  let datasets = [];
  let dataset = {};
  let arr = [];
  for (let i = 0; i < d.length; i++) {
    let data = d[i].d;
    if (typeof data === "object") {
      return {};
    }
    if (typeof data !== "string" && typeof data !== "number") {
      data = data.toString();
    }
    labelSet[data.toString()] = true;
    let numkeys = Object.keys(labelSet).length;
    if (numkeys > 100 || (i > 20 && numkeys / d.length > 0.5)) {
      return {};
    }
    arr.push({
      timeRange: [
        new Date(d[i].t * 1000),
        new Date(1000 * (d[i].t + (d[i].td === undefined ? 0 : d[i].td)))
      ],
      val: data
    });
  }
  if (dt < 1.5 * day) {
    return {
      timeline: {
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
                  data: arr
                }
              ]
            }
          ]
        }
      }
    };
  }
  // There are multiple days, so split the timeline by day
  let datas = [];
  let curdata = [];
  let i = 0;
  let curdp = arr[i];
  let curDate = curdp.timeRange[0];
  while (true) {
    if (datesAreOnSameDay(curDate, curdp.timeRange[0])) {
      // We add this datapoint to the current day array,
      if (datesAreOnSameDay(curDate, curdp.timeRange[1])) {
        // The entire datapoint fits in the same day, so just add it whole
        curdata.push(curdp);

        i += 1;
        if (i >= arr.length) {
          break;
        }
        curdp = arr[i];
      } else {
        // The datapoint does NOT fit in the day, so split it into a portion in the day, and a portion outside
        curdata.push({
          timeRange: [curdp.timeRange[0], endDate(curDate)],
          val: curdp.val
        });
        curdp = {
          timeRange: [nextDate(curDate), curdp.timeRange[1]],
          val: curdp.val
        };
      }
    } else {
      // Tranform the dates in curdata to all be the same day

      datas.push({
        label: moment(curDate).format("YYYY-MM-DD"),
        data: curdata
      });
      curdata = [];
      curDate = nextDate(curDate);
    }
  }

  // Now transform all the data timestamps to be the same day
  datas.forEach(cd =>
    cd.data.forEach(dp => {
      dp.timeRange[0].setFullYear(1970, 1, 2);
      dp.timeRange[1].setFullYear(1970, 1, 2);
    })
  );
  return {
    timeline: {
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

export default process;
