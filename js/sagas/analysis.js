import { delay } from 'redux-saga'
import { put, select, takeLatest } from 'redux-saga/effects'
import storage from '../storage';
import { cdbPromise } from '../util';

/**
 * Validates and queries ConnectorDB for a dataset. Shows errors if necessary.
 */
function* query(action) {
    // First, validate all fields
    let analysis = yield select((state) => state.pages.analysis);
    if (analysis.stream.length === 0 || analysis.stream.split("/").length != 3) {
        yield put({ type: "ANALYSIS_ERROR", value: "Invalid correlation stream name" });
        return;
    }

    // Do the same for each element of the dataset
    let datasetKeys = Object.keys(analysis.dataset);
    for (let i = 0; i < datasetKeys.length; i++) {
        let currentdataset = analysis.dataset[datasetKeys[i]]
        if (currentdataset.stream.length === 0 || currentdataset.stream.split("/").length != 3) {
            yield put({ type: "ANALYSIS_ERROR", value: "Invalid stream name (" + datasetKeys[i] + ")" });
            return;
        }
    }

    // Alright, validation complete. Let's query for the dataset
    yield put({ type: 'ANALYSIS_LOADING', value: true });

    let query = {
        posttransform: analysis.posttransform,
        dataset: analysis.dataset,
        t1: analysis.t1.unix(),
        t2: analysis.t2.unix(),
        limit: 0,
        stream: analysis.stream,
        allownil: false
    };
    try {
        let dataset = (yield cdbPromise(storage.cdb._doRequest("query/dataset", "POST", query), 5 * 60 * 1000));
        yield put({ type: "SHOW_DATASET", value: dataset });
    } catch (err) {
        console.log(err);
        yield put({ type: "ANALYSIS_ERROR", value: err.toString() });
    }
}

// Our watcher Saga: spawn a new incrementAsync task on each INCREMENT_ASYNC
export default function* analysisSaga() {
    yield takeLatest('DATASET_QUERY', query);
}