import { delay } from "redux-saga";
import { put, select, takeLatest } from "redux-saga/effects";
import moment from "moment";

import storage from "../storage";
import { cdbPromise } from "../util";

const day = 60 * 60 * 24;

// Given data, where we know the time range of the data is greater than the given time, it returns the data
// from the given time range. If the data is less than 5 datapoints, returns the last 5 datapoints anyways
function extract(data, time) {
  let t = data[data.length - 1].t;
  for (let i = data.length - 1; i >= 0; i--) {
    if (data[i].t < t - time) {
      // We really don't want to return less than 5 datapoints
      if (data.length - (i + 1) < 5) {
        return data.slice(data.length - 5);
      }
      return data.slice(i + 1);
    }
  }
  return data;
}

function* putData(path, data) {
  yield put({
    type: "STREAM_VIEW_SET",
    name: path,
    value: {
      bytime: true,
      t1: moment.unix(data[0].t),
      t2: moment().endOf("day"),
      firstvisit: false
    }
  });
  yield put({ type: "STREAM_VIEW_DATA", name: path, value: data });
}

// Query the last trange of data from the last datapoint
function* queryTimeData(uds, path, data, trange) {
  console.log("Querying", trange / day, "days of data");
  let t1 = moment.unix(data[data.length - 1].t - trange);
  let t2 = moment();
  try {
    yield put({
      type: "STREAM_VIEW_SET",
      name: path,
      value: { bytime: true, t1: t1, t2: t2, firstvisit: false }
    });
    data = yield cdbPromise(
      storage.cdb.timeStream(
        uds[0],
        uds[1],
        uds[2],
        t1.unix(),
        t2.unix(),
        20000
      ),
      20 * 1000
    );
    yield put({ type: "STREAM_VIEW_DATA", name: path, value: data });
  } catch (e) {
    console.log("Error getting data", e);
    yield put({
      type: "STREAM_VIEW_ERROR",
      name: path,
      value: { msg: e.toString() }
    });
    return;
  }
}

function* navigate(action) {
  let path = action.payload.pathname.slice(1);
  let uds = path.split("/");
  if (action.payload.hash !== "" || uds.length != 3 || uds[2].length === 0) {
    return;
  }
  // We just navigated to the streams page. Let's check if there is data already queried:
  let state = yield select(state => state.stream);

  if (state[path] !== undefined && !state[path].view.firstvisit) {
    // This page was already visited, so we don't query any data.
    return;
  }

  // We've got this - the page was now visited, and we're gonna query the data.
  yield put({
    type: "STREAM_VIEW_SET",
    name: path,
    value: { firstvisit: false }
  });

  // We are just getting the initial data from the stream.
  // First, let's query the last... say 100 datapoints.
  let data = [];
  try {
    data = yield cdbPromise(
      storage.cdb.indexStream(uds[0], uds[1], uds[2], -100, 0),
      5 * 1000
    );
  } catch (e) {
    yield put({
      type: "STREAM_VIEW_ERROR",
      name: path,
      value: { msg: e.toString() }
    });
    return;
  }

  // Now, let's check how these datapoints are on time. Our goal is to show a reasonable summary of the stream
  // The issue is that sometimes 100 datapoints is way too much, and other times, 100 datapoints is less than an hour of data.
  // A reasonable heuristic is how much time the datapoints represent. We assume that the data is fairly uniformly distributed
  // We want to display ~ 1 week of data, but don't want to display more than ~10000 datapoints, and also not less than ~10 datapoints

  let datarange = data.length > 0 ? data[data.length - 1].t - data[0].t : 0;

  if (data.length <= 15 || (datarange < 14 * day && data.length < 100)) {
    console.log("Showing entire dataset");
    // If there are <= 15 datapoints in the stream, just display them all.
    // Also if there are less than 2 weeks of data in the entire dataset, just show the full thing
    yield* putData(path, data);
    return;
  }
  if (datarange === 0) {
    console.log("No time range in last 100 datapoints.");
    yield* putData(path, data);
    return;
  }

  if (datarange > 14 * day) {
    console.log("Showing last 2 weeks of data");
    // There is more than 2 weeks of data. Let's extract the last two weeks and display that
    yield* putData(path, extract(data, 14 * day));
    return;
  }
  if (datarange > 4 * day) {
    console.log("Showing last 100 datapoints");
    // There is more than 2 weeks of data. Let's extract the last two weeks and display that
    yield* putData(path, data);
    return;
  }

  // The data range is less than 4 days. We really want to see a larger time range. The only question now is
  // how big we expect the stream data to get. We now find the expected number of datapoints per day
  let perday = 100 / datarange * day;

  if (perday <= 500) {
    // Nice, let's show a week of data
    yield* queryTimeData(uds, path, data, day * 7);
    return;
  }
  if (perday <= 1000) {
    // 4 days should do it
    yield* queryTimeData(uds, path, data, day * 4);
    return;
  }
  if (perday <= 2000) {
    // 2 days should do it
    yield* queryTimeData(uds, path, data, day * 2);
    return;
  }

  if (perday <= 5000) {
    yield* queryTimeData(uds, path, data, day);
    return;
  }

  // Oh crap. There might be more than 8k per day. Let's just display 1000 of them
  console.log("Stream looks dense. Querying last 1000 datapoints.");
  yield put({
    type: "STREAM_VIEW_SET",
    name: path,
    value: { bytime: false, i1: -1000, i2: 0, firstvisit: false }
  });
  try {
    data = yield cdbPromise(
      storage.cdb.indexStream(uds[0], uds[1], uds[2], -1000, 0),
      15 * 1000
    );
  } catch (e) {
    yield put({
      type: "STREAM_VIEW_ERROR",
      name: path,
      value: { msg: e.toString() }
    });
    return;
  }

  yield put({ type: "STREAM_VIEW_DATA", name: path, value: data });
  return;
}

export default function* downlinkSaga() {
  yield takeLatest("@@router/LOCATION_CHANGE", navigate);
}
