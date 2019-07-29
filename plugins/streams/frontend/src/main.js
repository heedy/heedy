import Create from "./main/create.vue";
import Stream from "./main/stream.vue";

function setup(app) {

  if (app.info.user!=null) {
    app.source.addCreator({
      key: "stars",
      text: "Star Rating",
      icon: "star",
      route: "/create/source/stream/stars"
    });

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

  app.source.typePath("stream","stream/");
  app.source.addRoute("stream",{
    path: "/",
    component: Stream
  });
}

export default setup;
