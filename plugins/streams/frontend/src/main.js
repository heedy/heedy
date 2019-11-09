import Create from "./main/create.vue";
import Views from "./main/views.vue";
import Header from "./main/header.vue";
import vuexModule from "./main/vuex.js";
import StreamInjector from "./main/injector";
import Update from "./main/update.vue";

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

  app.worker.import("streams/worker.mjs");

  app.source.addComponent({
    component: Views,
    type: "stream",
    key: "views",
    weight: 5
  });

  app.streams.addView("datatable", () => import("./views/datatable.mjs"));
  app.streams.addView("insert", () => import("./views/insert.mjs"));
  app.streams.addView("apexchart", () => import("./views/apexchart.mjs"));


  app.source.addComponent({
    component: Header,
    type: "stream",
    key: "header"
  });

  app.source.addType({
    type: "stream",
    title: "Stream",
    list_title: "Streams",
    icon: "timeline"
  });


  //app.source.replacePage("stream", Stream);
}

export default setup;