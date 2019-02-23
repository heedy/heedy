import VueRouter from "vue-router";

import Theme from "./components/Theme.vue";
import Basics from "./components/Basics.vue";

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
