import Create from "./main/create.vue";
import Views from "./main/views.vue";
import Header from "./main/header.vue";
import vuexModule from "./main/vuex.js";
import TimeseriesInjector from "./main/injector";
import Update from "./main/update.vue";

function setup(app) {
  app.store.registerModule("timeseries", vuexModule);
  app.inject("timeseries", new TimeseriesInjector(app));

  if (app.info.user != null) {
    app.object.addCreator({
      key: "rawtimeseries",
      title: "Timeseries",
      description: "Manually gather data.",
      icon: "timeline",
      route: "/create/object/timeseries"
    });

    app.object.addRoute({
      path: "/timeseries/update",
      component: Update
    });

    app.addRoute({
      path: "/create/object/timeseries",
      component: Create
    });
  }

  app.worker.import("timeseries/worker.mjs");

  app.object.addComponent({
    component: Views,
    type: "timeseries",
    key: "views",
    weight: 5
  });

  app.timeseries.addView("datatable", () => import("./views/datatable.mjs"));
  app.timeseries.addView("insert", () => import("./views/insert.mjs"));
  app.timeseries.addView("apexchart", () => import("./views/apexchart.mjs"));
  app.timeseries.addView("chartjs", () => import("./views/chartjs.mjs"));
  app.timeseries.addView("timeline", () => import("./views/timeline.mjs"));
  app.timeseries.addView("horizon", () => import("./views/horizon.mjs"));

  app.object.addComponent({
    component: Header,
    type: "timeseries",
    key: "header"
  });

  app.object.addType({
    type: "timeseries",
    title: "Timeseries",
    list_title: "Timeseries",
    icon: "timeline"
  });

  //app.object.replacePage("timeseries", Timeseries);
}

export default setup;
