class App {
  constructor(appinfo) {
    this.appinfo = appinfo;
  }
  hi() {
    console.log("hi");
  }
}

var app = new App(appinfo);

async function setup() {
  console.log("Setting up...");

  console.log("Importing plugins");

  let plugins = await Promise.all(
    appinfo.frontend.map(f => import("./" + f.path))
  );

  for (let i = 0; i < plugins.length; i++) {
    console.log("Preparing", appinfo.frontend[i].name);
    plugins[i].default(app);
  }
}

setup();
