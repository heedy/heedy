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

  // Start running the import statements
  let plugins = appinfo.frontend.map(f => import("./" + f.path));

  for (let i = 0; i < plugins.length; i++) {
    console.log("Preparing", appinfo.frontend[i].name);
    (await plugins[i]).default(app);
  }
}

setup();
