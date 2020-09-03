var objectRoutesMap = {};
var objectRoutes = [];

class ObjectInjector {
  /**
   * Deals with objects
   * @param {*} frontend
   */
  constructor(frontend) {
    this.store = frontend.store;

    let queryObject = (e) => {
      if (
        this.store.state.heedy.objects[e.object] !== undefined ||
        this.store.state.heedy.userObjects[e.user] !== undefined ||
        (e.app !== undefined &&
          this.store.state.heedy.appObjects[e.app] !== undefined)
      ) {
        this.store.dispatch("readObject_", {
          id: e.object,
        });
      }
    };
    // Subscribe to all object events, so that the object list
    // can be kept up-to-date
    if (frontend.info.user != null) {
      frontend.websocket.subscribe(
        "object_create",
        {
          event: "object_create",
          user: frontend.info.user.username,
        },
        queryObject
      );
      frontend.websocket.subscribe(
        "object_update",
        {
          event: "object_update",
          user: frontend.info.user.username,
        },
        queryObject
      );
      frontend.websocket.subscribe(
        "object_delete",
        {
          event: "object_delete",
          user: frontend.info.user.username,
        },
        (e) => {
          if (
            this.store.state.heedy.objects[e.object] !== undefined ||
            this.store.state.heedy.userObjects[e.user] !== undefined ||
            (e.app !== undefined &&
              this.store.state.heedy.appObjects[e.app] !== undefined)
          ) {
            this.store.commit("setObject", {
              id: e.object,
              isNull: true,
            });
          }
        }
      );
    }

    // Respond to object queries from the worker
    frontend.worker.addHandler("get_object", (ctx, msg) =>
      this.get(msg.id).then((obj) =>
        frontend.worker.postMessage("get_object", obj)
      )
    );
  }

  /**
   * Returns a promise that gets the given object
   * @param {*} oid Object ID to get
   */
  get(oid) {
    return new Promise((resolve, reject) => {
      this.store.dispatch("readObject", {
        id: oid,
        callback: () => resolve(this.store.state.heedy.objects[oid]),
      });
    });
  }

  /**
   * Identical to a menu item, it is displayed in a special object creation menu
   * @param {*} c The creator to add
   */
  addCreator(c) {
    this.store.commit("addObjectCreator", c);
  }
  addComponent(c) {
    this.store.commit("addObjectComponent", c);
  }
  /**
   * Adds or overwrites an object type. It is passed a dictionary of things to set.
   * @param {*} c
   */
  setType(c) {
    this.store.commit("setObjectType", c);
  }

  /**
   * Adds a route to objects. It
   * automatically takes /object/:objectid/{r.path}.
   * If the route works only on a specific object type, it is recommended to
   * prefix it with the type, ie: r.path = /timeseries/...
   * @param {*} r
   */
  addRoute(r) {
    objectRoutesMap[r.path] = r;
  }

  $onInit() {
    // Need to set the objectRoutes with the right values:
    Object.values(objectRoutesMap).reduce((_, r) => {
      if (r.path.startsWith("/")) {
        r.path = r.path.substring(1, r.path.length);
      }
      objectRoutes.push(r);
      return null;
    }, null);
  }
}

export { objectRoutes };
export default ObjectInjector;
