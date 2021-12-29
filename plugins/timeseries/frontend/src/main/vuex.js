import Vue from "../../dist/vue.mjs";
// import api from "../../util.mjs";

export default {
  state: {
    visualizations: {},
    types: {}
  },
  mutations: {
    addTSVisualization(state, v) {
      Vue.set(state.visualizations, v.key, v.component);
    },
    addTSType(state, v) {
      Vue.set(state.types, v.key, v);
    }
  },
};
