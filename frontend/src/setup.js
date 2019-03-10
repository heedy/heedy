import VueRouter from "vue-router";

import Theme from "./embedded/setup/Theme.vue";
import Basics from "./embedded/setup/Basics.vue";

Vue.use(VueRouter);

const router = new VueRouter({
  routes: [
    {
      path: "/",
      name: "Basics",
      component: Basics
    }
  ]
});

new Vue({
  router,
  render: h => h(Theme)
}).$mount("#app");
