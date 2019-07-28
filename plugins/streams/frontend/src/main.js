import Create from "./main/create.vue";

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
}

export default setup;
