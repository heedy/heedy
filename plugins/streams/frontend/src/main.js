import Create from "./main/create.vue";
import DataVis from "./main/datavis.vue";
import Header from "./main/header.vue";
import vuexModule from "./main/vuex.js";
import StreamInjector from "./main/injector";
import Update from "./main/update.vue";

import DataTable from "./main/visualizations/datatable.vue";
import Insert from "./main/visualizations/insert.vue";

function setup(app) {

  app.store.registerModule("streams", vuexModule);
  app.inject("streams", new StreamInjector(app));

  if (app.info.user != null) {

    app.source.addCreator({
      key: "rawstream",
      text: "Stream",
      icon: "timeline",
      route: "/create/source/stream"
    });

    app.source.addRoute({
      path: "/stream/update",
      component: Update
    });

    app.addRoute({
      path: "/create/source/stream",
      component: Create
    });
  }

  app.worker.add("streams/worker.mjs");

  app.source.addComponent({
    component: DataVis,
    type: "stream",
    key: "visualization",
    weight: 5
  });

  app.streams.addVisualization("datatable", DataTable);
  app.streams.addVisualization("insert", Insert);


  app.source.addComponent({
    component: Header,
    type: "stream",
    key: "header"
  });


  //app.source.replacePage("stream", Stream);
}

export default setup;