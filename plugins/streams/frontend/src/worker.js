import api from "../api.mjs";

import StreamInjector from "./worker/injector.js";

import datatable from "./worker/processors/datatable.js";
import insert from "./worker/processors/insert.js";
import linechart from "./worker/processors/linechart.js";

function setup(wkr) {
    console.log("stream_worker: starting");

    wkr.inject("streams", new StreamInjector(wkr));

    wkr.streams.addProcessor("datatable", datatable);
    wkr.streams.addProcessor("insert", insert);
    wkr.streams.addProcessor("linechart", linechart);
}

export default setup;