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

    // Do the same for each element of the dataset
    let datasetKeys = Object.keys(analysis.dataset);
    for (let i = 0; i < datasetKeys.length; i++) {
        let currentdataset = analysis.dataset[datasetKeys[i]]
        if (currentdataset.stream.length === 0 || currentdataset.stream.split("/").length != 3) {
            yield put({ type: "ANALYSIS_ERROR", value: "Invalid stream name (" + datasetKeys[i] + ")" });
            return;
        }
    }

    // This is the format in which the ConnectorDB server expects a query
    let query = {
        posttransform: analysis.posttransform,
        dataset: analysis.dataset,
        t1: analysis.t1.unix(),
        t2: analysis.t2.unix(),
        limit: 0,
        allownil: false
    };

    if (analysis.xdataset) {
        if (analysis.stream.length === 0 || analysis.stream.split("/").length != 3) {
            yield put({ type: "ANALYSIS_ERROR", value: "Invalid correlation stream name" });
            return;
        }
        query.stream = analysis.stream;
        query.transform = analysis.transform;
    } else {
        // Make sure that dt is a number
        let dt = parseFloat(analysis.dt);
        if (isNaN(dt) || dt < 0.001) {
            yield put({ type: "ANALYSIS_ERROR", value: "Invalid time delta (" + analysis.dt + ")" });
            return;
        }
        query.dt = dt;
    }

    // Alright, validation complete. Let's query for the dataset
    yield put({ type: "SHOW_DATASET", value: [] });    // First clear any data that might be there
    yield put({ type: 'ANALYSIS_LOADING', value: true });   // Next turn loading on



    try {
        let dataset = (yield cdbPromise(storage.cdb._doRequest("query/dataset", "POST", query), 5 * 60 * 1000));
        yield put({ type: "SHOW_DATASET", value: dataset });
    } catch (err) {
        console.log(err);
        yield put({ type: "ANALYSIS_ERROR", value: err.toString() });
        yield put({ type: 'ANALYSIS_LOADING', value: false });
    }
}

import React from 'react';
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/monokai.css';
import CodeMirror from 'react-codemirror';
import 'codemirror/mode/python/python';

function* showPython(action) {
    let analysis = yield select((state) => state.pages.analysis);
    let username = yield select((state) => state.site.thisUser.name);

    let interpolations = ""
    let datasetKeys = Object.keys(analysis.dataset);
    for (let i = 0; i < datasetKeys.length; i++) {
        let currentdataset = analysis.dataset[datasetKeys[i]]
        interpolations = interpolations + `d.addStream("${currentdataset.stream}","${currentdataset.interpolator}",`
            + (currentdataset.transform != "" ? `transform=${JSON.stringify(currentdataset.transform)},` : "")
            + `colname="${datasetKeys[i]}")`
            + "\n";
    }

    let pythoncode = `import connectordb
from connectordb.query import Dataset

import getpass
p = getpass.getpass()

cdb = connectordb.ConnectorDB("${username}",p,url="${SiteURL}")

d = Dataset(cdb,"${analysis.stream}",t1=${analysis.t1.unix()},t2=${analysis.t2.unix()}`
        + (analysis.transform == "" ? "" : `,
        transform=${JSON.stringify(analysis.transform)}`)
        + (analysis.posttransform == "" ? "" : `,
        posttransform=${JSON.stringify(analysis.posttransform)}`)
        + ")\n"
        + interpolations
        + `
data = d.run()
`;

    yield put({
        type: "SHOW_DIALOG",
        value: {
            title: "Python Code",
            open: true,
            contents: (
                <div>
                    <CodeMirror value={pythoncode} options={{
                        mode: "text/x-python",
                        lineWrapping: true,
                        readOnly: true
                    }} />
                    <a style={{ float: "right" }} href="http://connectordb-python.readthedocs.io/en/latest/">Python API Docs</a>
                </div>
            )
        }
    });
}

// This should move to a different file at some point
function* showPythonQuery(action) {
    let stream = yield select((state) => state.stream[action.value].view);
    let username = yield select((state) => state.site.thisUser.name);
    let varname = action.value.split("/")[2];
    let pythoncode = `import connectordb

import getpass
p = getpass.getpass()

cdb = connectordb.ConnectorDB("${username}",p,url="${SiteURL}")

${varname} = cdb("${action.value}")

`
    if (!stream.bytime && stream.transform == "") {
        pythoncode += `data = ${varname}[${stream.i1}:${stream.i2}]`;
    } else if (!stream.bytime && stream.transform != "") {
        pythoncode += `data = ${varname}(i1=${stream.i1},i2=${stream.i2},transform=${JSON.stringify(stream.transform)})`;
    } else {
        // by time
        pythoncode += `data = ${varname}(t1=${stream.t1.unix()},t2=${stream.t2.unix()}`
            + (stream.transform != "" ? `,transform=${JSON.stringify(stream.transform)}` : "")
            + ")";
    }

    yield put({
        type: "SHOW_DIALOG",
        value: {
            title: "Python Code",
            open: true,
            contents: (
                <div>
                    <CodeMirror value={pythoncode} options={{
                        mode: "text/x-python",
                        lineWrapping: true,
                        readOnly: true
                    }} />
                    <a style={{ float: "right" }} href="http://connectordb-python.readthedocs.io/en/latest/">Python API Docs</a>
                </div>
            )
        }
    });

}

// Our watcher Saga: spawn a new incrementAsync task on each INCREMENT_ASYNC
export default function* analysisSaga() {
    yield takeLatest('DATASET_QUERY', query);
    yield takeLatest('SHOW_ANALYSIS_CODE', showPython);
    yield takeLatest('SHOW_QUERY_CODE', showPythonQuery);
}