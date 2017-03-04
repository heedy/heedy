import { delay } from 'redux-saga'
import { put, takeLatest } from 'redux-saga/effects'

// We might want to do certain actions on navigation
function* navigate(action) {
    yield put({ type: 'SET_ERROR_VALUE', value: action.value });
    yield delay(5000);
    yield put({ type: 'SET_ERROR_VALUE', value: null });
}

// Our watcher Saga: spawn a new incrementAsync task on each INCREMENT_ASYNC
export default function* downlinkSaga() {
    //yield takeLatest('@@router/LOCATION_CHANGE', navigate);
}