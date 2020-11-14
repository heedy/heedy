import Create from "./main/create.vue";
import VisTimeseries from "./main/vis_timeseries.vue";
import DatasetVisualization from "./main/dataset_visualization.vue";
import Header from "./main/header.vue";
import vuexModule from "./main/vuex.js";
import TimeseriesInjector from "./main/injector";
import Update from "./main/update.vue";
import Dataset from "./main/dataset/editor.vue";
import InputPage from "./main/inserter/inputpage.vue";

import RatingInserter from "./main/inserter/custom/rating.vue";

function setup(frontend) {
  frontend.store.registerModule("timeseries", vuexModule);
  frontend.inject("timeseries", new TimeseriesInjector(frontend));

  // The component that permits visualizing a dataset
  frontend.vue.component("h-dataset-visualization", DatasetVisualization);

  // Add the default timeseries types
  frontend.timeseries.addType({
    key: "number",
    schema: {
      type: "number"
    },
    icon: "timeline",
    title: "Number"
  });
  frontend.timeseries.addType({
    key: "string",
    schema: {
      type: "string"
    },
    icon: "fas fa-list",
    title: "String"
  });
  frontend.timeseries.addType({
    key: "rating",
    schema: {
      type: "integer",
      minimum: 0,
      maximum: 10,
      "x-display": "custom-rating"
    },
    icon: "star",
    title: "Rating"
  });
  frontend.timeseries.addCustomInserter("rating", RatingInserter);

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
      key: "insert",
      text: "Manual Inputs",
      icon: "fas fa-star",
      route: "/timeseries/insert"
    });

    frontend.addMenuItem({
      key: "dataset",
      text: "Data Analysis",
      icon: "fas fa-chart-bar",
      route: "/timeseries/dataset",
    });
  }

  frontend.addRoute({
    path: "/timeseries/dataset",
    component: Dataset,
  });
  frontend.addRoute({
    path: "/timeseries/insert",
    component: InputPage,
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
  frontend.timeseries.addVisualization("timeline", () =>
    import("./visualizations/timeline.mjs")
  );
  /*
  
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
