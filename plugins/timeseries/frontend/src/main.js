import Create from "./main/create.vue";
import Views from "./main/views.vue";
import Header from "./main/header.vue";
import vuexModule from "./main/vuex.js";
import TimeseriesInjector from "./main/injector";
import Update from "./main/update.vue";

function setup(frontend) {
  frontend.store.registerModule("timeseries", vuexModule);
  frontend.inject("timeseries", new TimeseriesInjector(frontend));

  if (frontend.info.user != null) {
    frontend.objects.addCreator({
      key: "rawtimeseries",
      title: "Timeseries",
      description: "Manually gather data.",
      icon: "timeline",
      route: "/create/object/timeseries",
    });

    frontend.addRoute({
      path: "/create/object/timeseries",
      component: Create,
    });
  }

  frontend.worker.import("timeseries/worker.mjs");

  frontend.objects.addComponent({
    component: Views,
    type: "timeseries",
    key: "body",
    weight: 5,
  });

  frontend.timeseries.addView("datatable", () =>
    import("./views/datatable.mjs")
  );
  frontend.timeseries.addView("insert", () => import("./views/insert.mjs"));
  frontend.timeseries.addView("chartjs", () => import("./views/chartjs.mjs"));
  frontend.timeseries.addView("timeline", () => import("./views/timeline.mjs"));
  frontend.timeseries.addView("horizon", () => import("./views/horizon.mjs"));

  frontend.objects.addComponent({
    component: Header,
    type: "timeseries",
    key: "header",
  });

  frontend.objects.setType({
    type: "timeseries",
    title: "Timeseries",
    list_title: "Timeseries",
    icon: "timeline",
    update: Update,
  });
}

export default setup;
