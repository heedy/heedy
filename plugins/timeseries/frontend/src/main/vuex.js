import Vue from "../../dist/vue.mjs";
// import api from "../../rest.mjs";

export default {
  state: {
    visualizations: {},
  },
  mutations: {
    addTSVisualization(state, v) {
      Vue.set(state.visualizations, v.key, v.component);
    },
  },
};
