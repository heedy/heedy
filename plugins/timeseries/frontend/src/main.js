import Create from "./main/create.vue";
import VisTimeseries from "./main/vis_timeseries.vue";
import DatasetVisualization from "./main/dataset_visualization.vue";
import Header from "./main/header.vue";
import vuexModule from "./main/vuex.js";
import TimeseriesInjector from "./main/injector";
import Update from "./main/update.vue";
import Dataset from "./main/dataset/editor.vue";
import InputPage from "./main/inputpage.vue";
import SchemaEditor from "./main/components/schema_editor.vue";
import RangePicker from "./main/components/range_picker.vue";
import DatasetToolbarItem from "./main/components/dataset_toolbar_item.vue";
import DatapointInserter from "./main/components/datapoint_inserter.vue";
import DataTable from "./main/components/datatable.vue";
import BasicTable from "./main/components/table.vue";
import DurationEditor from "./main/components/duration.vue";
import DataTableVisualization from "./visualizations/datatable.vue";
import BasicTableVisualization from "./visualizations/table.vue";

import RatingInserter from "./main/components/schema/rating.vue";
import EnumEditor from "./main/components/enum.vue";

import VCalendar from "../dist/v-calendar.mjs";

function setup(frontend) {
  frontend.store.registerModule("timeseries", vuexModule);
  frontend.inject("timeseries", new TimeseriesInjector(frontend));

  frontend.vue.use(VCalendar, {
    componentPrefix: 'vc'
  });

  // The component that permits visualizing a dataset
  frontend.vue.component("h-dataset-visualization", DatasetVisualization);
  frontend.vue.component("h-schema-editor", SchemaEditor);
  frontend.vue.component("h-timeseries-datapoint-inserter", DatapointInserter);
  frontend.vue.component("h-timeseries-range-picker", RangePicker);
  frontend.vue.component("h-timeseries-datatable", DataTable);
  frontend.vue.component("h-table", BasicTable);
  frontend.vue.component("h-duration-editor", DurationEditor);

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
    title: "Star Rating"
  });
  frontend.addSchemaFormElement("rating", RatingInserter);
  frontend.timeseries.addType({
    key: "enum",
    schema: {
      type: "string",
      enum: ["my_event"]
    },
    meta: {
      type: "object",
      properties: {
        type: { type: "string", enum: ["string"] },
        enum: { type: "array", items: { type: "string" } }
      }
    },
    editor: EnumEditor,
    icon: "star",
    title: "Events"
  });

  if (frontend.info.user != null) {
    frontend.objects.addCreator({
      key: "rawtimeseries",
      title: "Timeseries",
      description: "Manually gather data.",
      icon: "timeline",
      route: "/create/object/timeseries",
    });

    frontend.addRoute({
      path: "/create/object/timeseries/:datatype?",
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

    frontend.addRoute({
      path: "/timeseries/dataset",
      component: Dataset,
    });
    frontend.addRoute({
      path: "/timeseries/insert",
      component: InputPage,
    });

  }





  frontend.worker.import("timeseries/worker.mjs");

  frontend.objects.addComponent({
    component: VisTimeseries,
    type: "timeseries",
    key: "body",
  });
  frontend.timeseries.addVisualization("chartjs", () =>
    import("./visualizations/chartjs.mjs")
  );
  frontend.timeseries.addVisualization("timeline", () =>
    import("./visualizations/timeline.mjs")
  );
  // The data table/basic table doesn't use any external libraries, so we can just import it
  frontend.timeseries.addVisualization("datatable", DataTableVisualization);
  frontend.timeseries.addVisualization("table", BasicTableVisualization);
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

  frontend.objects.addMenu((o) => {
    if (o.type != "timeseries") {
      return {};
    }
    return {
      analysis: {
        toolbar_component: DatasetToolbarItem,
        menu_component: DatasetToolbarItem,
        toolbar_props: { objectid: o.id },
        menu_props: { objectid: o.id, isList: true },
        toolbar: true,
        weight: -0.5
      }
    }
  })
}

export default setup;
