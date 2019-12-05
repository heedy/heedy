import Vue from "vue";
import Vuex from "vuex";

import Auth from "./auth/main.vue";

// store is a global variable.
export const store = new Vuex.Store({
  state: authinfo
});

// Vue is used as a global
export const vue = new Vue({
  store: store,
  render: h => h(Auth)
});

// Mount it
vue.$mount("#app");