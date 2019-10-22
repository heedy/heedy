import Vue, {
  VueRouter,
  Vuetify
} from "./dist/vue.mjs";

import Theme from "./embedded/setup/Theme.vue";
import Create from "./embedded/setup/Create.vue";

Vue.use(VueRouter);

const router = new VueRouter({
  routes: [{
    path: "/",
    name: "Create",
    component: Create
  }]
});
const vuetify = new Vuetify({
  icons: {
    iconfont: 'md',
  },
});

new Vue({
  router,
  vuetify: vuetify,
  render: h => h(Theme)
}).$mount("#app");