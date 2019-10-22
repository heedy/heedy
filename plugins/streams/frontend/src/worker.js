import api from "../api.mjs";

import StreamInjector from "./worker/injector.js";

import datatable from "./worker/processors/datatable.js";
import insert from "./worker/processors/insert.js";

function setup(wkr) {
    console.log("Streams worker");

    wkr.inject("streams", new StreamInjector(wkr));

    wkr.streams.addProcessor("datatable", datatable);
    wkr.streams.addProcessor("insert", insert);
}

export default setup;