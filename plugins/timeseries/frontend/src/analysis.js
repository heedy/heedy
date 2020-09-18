function cleanDT(ts) {
  // TODO: This is actually a bug in the underlying heedy code
  for (let i = 0; i < ts.length - 1; i++) {
    if (ts[i].dt !== undefined && ts[i].t + ts[i].dt > ts[i + 1].t) {
      ts[i].dt = ts[i + 1].t - ts[i].t;
    }
  }
  return ts;
}

/**
 * Splits datapoints with durations into two elements - one at start of the duration,
 * and one at end of the duration
 *
 * @param {*} ts timeseries
 */
function explicitDuration(ts, offset = 0.001) {
  let res = new Array(ts.length * 2);
  let j = 0;
  for (let i = 0; i < ts.length; i++) {
    res[j] = ts[i];
    j++;
    if (ts[i].dt !== undefined && ts[i].dt != 0) {
      res[j] = {
        t: ts[i].t + ts[i].dt - offset,
        d: ts[i].d,
      };
      j++;
    }
  }
  return res.slice(0, j);
}

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

function isNumeric(ts) {
  return ts.every((dp) => !isNaN(dp.d));
}

function isBoolean(ts) {
  return ts.every((dp) => typeof dp.d === "boolean");
}

function isObject(ts) {
  return ts.every((dp) => typeof dp.d === "object");
}

function getType(ts) {
  if (ts.length == 0) {
    return "";
  }
  let curtype = typeof ts[0].d;
  if (ts.every((dp) => typeof dp.d === curtype)) {
    return curtype;
  }
  return "";
}

function getKeys(ts) {
  let vals = {};
  ts.forEach((dp) => {
    Object.keys(dp.d).forEach((k) => {
      if (vals[k] === undefined) {
        vals[k] = 0;
      }
      vals[k]++;
    });
  });
  return vals;
}

/*
function getType(ts) {
  if (isNumeric(ts)) {
    return "number";
  }
  if (ts.every((dp) => typeof dp.d === "string")) {
    // Check if it is categorical
    let vals = {};
    ts.forEach((dp) => {
      vals[dp.d] = true;
    });
    let keynum = Object.keys(vals).length;
    if (keynum < 100 && keynum < ts.length / 3) {
      return "categorical";
    }

    return "string";
  }
  return null;
}*/

export {
  perDay,
  explicitDuration,
  isNumeric,
  day,
  getType,
  cleanDT,
  isBoolean,
  getKeys,
};
