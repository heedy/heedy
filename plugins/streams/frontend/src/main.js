import Create from "./main/create.vue";
import StreamHeader from "./main/stream_header.vue";

function setup(app) {

  if (app.info.user != null) {

    app.source.addCreator({
      key: "rawstream",
      text: "Stream",
      icon: "timeline",
      route: "/create/source/stream"
    });

    app.addRoute({
      path: "/create/source/stream",
      component: Create
    });
  }
  /* Will want to have a better replacement header once 
  start adding stream visualizations.
  app.source.addComponent({
    component: StreamHeader,
    type: "stream",
    key: "header"
  })
  */
  //app.source.replacePage("stream", Stream);
}

export default setup;