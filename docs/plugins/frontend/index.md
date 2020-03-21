# Frontend

A plugin performs all of its modifications to the heedy frontend through the use of
a single javascript module. Heedy is made aware of this module in the plugin's `heedy.conf`, by setting the `frontend` variable in the plugin's block. The variable is set to the module's path relative to the `public/static` directory in the plugin's folder.

```javascript
plugin "myplugin" {
    frontend = "myplugin/main.mjs"
}
```

While not strictly required, it is recommended that you prefix the module path with your plugin name (as was done here). In this case, heedy's frontend code will look for a javascript module file served at `/static/myplugin/main.mjs`. Then, for heedy's backend to serve the file at the correct location, you need to create a `main.mjs` file, and place it at `public/static/myplugin/main.mjs` in your plugin's folder. The minimal content of your main.mjs is the following:

```javascript
function setup(frontend) {
  // Initialize your plugin's frontend here
}
export default setup;
```

## The `frontend` object

Using the `frontend` object passed into your module's setup function, you can tell heedy's UI how to accommodate your plugin. The UI is built with [Vue](https://vuejs.org/), so the frontend object allows you to directly deal with Vue components:

```javascript
function setup(frontend) {
  // Register a vuex store for your plugin (optional)
  frontend.store.registerModule("myplugin", vuexModule);
  // Add a route to the ui, which will be accessibe from `/#/myplugin/myroute`
  frontend.addRoute({
    path: "/myplugin/myroute",
    component: MyComponent
  });
  // Add an item to the main menu that will redirect to the registered route
  frontend.addMenuItem({
    key: "mypluginMenuItem",
    text: "My Plugin",
    icon: "home",
    route: "/myplugin/myroute",
    location: "primary"
  });
}
export default setup;
```

### Injected Functionality

Each heedy plugin can attach additional functionality to the frontend object (i.e. inject their own objects into the frontend). In a bare heedy install, the following registered classes extend the frontend's functionality:

```eval_rst
.. toctree::
    :maxdepth: 1

    users
    apps
    objects
    settings
    websocket
    worker
    timeseries
```

Each of these can be accessed as a property of the `frontend` object (ex: `frontend.websocket`, `frontend.worker`, `frontend.timeseries`).

### Frontend API

```eval_rst
.. js:autoclass:: Frontend
    :members:
```
