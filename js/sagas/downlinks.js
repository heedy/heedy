import { delay } from 'redux-saga'
import { put, takeLatest } from 'redux-saga/effects'

function* showError(action) {
    yield put({ type: 'SET_ERROR_VALUE', value: action.value });
    yield delay(5000);
    yield put({ type: 'SET_ERROR_VALUE', value: null });
}

// Our watcher Saga: spawn a new incrementAsync task on each INCREMENT_ASYNC
export default function* downlinkSaga() {
    yield takeLatest('SHOW_ERROR', showError);
}