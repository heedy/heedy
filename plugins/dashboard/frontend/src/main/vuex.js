import Vue from "../../dist/vue.mjs";

export default {
  state: {
    types: {},
  },
  mutations: {
    addDashboardType(state, v) {
      Vue.set(state.types, v.type, v.component);
    },
  },
};
