import Create from "./main/create.vue";
import VisTimeseries from "./main/vis_timeseries.vue";
import DatasetVisualization from "./main/dataset_visualization.vue";
import Header from "./main/header.vue";
import vuexModule from "./main/vuex.js";
import TimeseriesInjector from "./main/injector";
import Update from "./main/update.vue";
import Dataset from "./main/dataset/editor.vue";

function setup(frontend) {
  frontend.store.registerModule("timeseries", vuexModule);
  frontend.inject("timeseries", new TimeseriesInjector(frontend));

  // The component that permits visualizing a dataset
  frontend.vue.component("h-dataset-visualization", DatasetVisualization);

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

    frontend.addMenuItem({
      key: "dataset",
      text: "Data Analysis",
      icon: "fas fa-chart-bar",
      route: "/dataset",
    });
  }

  frontend.addRoute({
    path: "/dataset",
    component: Dataset,
  });

  frontend.worker.import("timeseries/worker.mjs");

  frontend.objects.addComponent({
    component: VisTimeseries,
    type: "timeseries",
    key: "body",
  });

  frontend.timeseries.addVisualization("table", () =>
    import("./visualizations/table.mjs")
  );
  frontend.timeseries.addVisualization("chartjs", () =>
    import("./visualizations/chartjs.mjs")
  );
  /*
  frontend.timeseries.addVisualization("insert", () =>
    import("./visualizations/insert.mjs")
  );
  frontend.timeseries.addVisualization("chartjs", () =>
    import("./visualizations/chartjs.mjs")
  );
  frontend.timeseries.addVisualization("timeline", () =>
    import("./visualizations/timeline.mjs")
  );
  frontend.timeseries.addVisualization("horizon", () =>
    import("./visualizations/horizon.mjs")
  );
  */

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
