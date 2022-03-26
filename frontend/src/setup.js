import Vue, {
  VueRouter,
  Vuetify
} from "./dist/vue.mjs";

import Theme from "./setup/Theme.vue";
import Create from "./setup/Create.vue";

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

// Remove any prior serviceworkers, so that heedy gets fresh content
if ("serviceWorker" in navigator) {
  navigator.serviceWorker.getRegistrations().then( (registrations) =>{ 
    for (let registration of registrations) { 
      registration.unregister(); 
    } 
  });
}