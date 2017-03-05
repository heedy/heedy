import { delay } from 'redux-saga'
import { put, select, takeLatest } from 'redux-saga/effects'

import storage from '../storage';
import { cdbPromise } from '../util';


function* navigate(action) {
    if (action.payload.hash !== "#downlinks" || action.payload.pathname !== "/") {
        return;
    }
    // We are to navigate to the downlinks page. Let's refresh the downlink stream list
    let username = yield select((state) => state.site.thisUser.name);
    try {
        let streams = (yield cdbPromise(storage.cdb.listUserStreams(username, "*", false, true, true))); //.map((s) => ({ ...s, schema: JSON.parse(s.schema) }));
        yield put({ type: 'UPDATE_DOWNLINKS', value: streams });
    } catch (err) {
        console.log(err);
        yield put({ type: "SHOW_STATUS", value: err.toString() });
    }
}


export default function* downlinkSaga() {
    yield takeLatest('@@router/LOCATION_CHANGE', navigate);
}