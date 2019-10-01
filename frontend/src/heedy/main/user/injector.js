var userRoutesMap = {};
var userRoutes = [];
class User {
    constructor(store) {
        this.store = store;
    }

    addRoute(r) {
        userRoutesMap[r.path] = r;
    }

    addComponent(c) {
        this.store.commit("addUserComponent", c);
    }

    $onInit() {
        Object.values(userRoutesMap).reduce((_, r) => {
            if (r.path.startsWith("/")) {
                r.path = r.path.substring(1, r.path.length);
            }
            userRoutes.push(r);
            return null;
        }, null);
    }
}

export {
    userRoutes
}
export default User;