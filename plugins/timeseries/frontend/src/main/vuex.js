import Vue from "../../dist/vue.mjs";
// import api from "../../rest.mjs";

export default {
  state: {
    visualizations: {},
    customInserters: {},
    types: {}
  },
  mutations: {
    addTSVisualization(state, v) {
      Vue.set(state.visualizations, v.key, v.component);
    },
    addTSCustomInserter(state, v) {
      Vue.set(state.customInserters, v.key, v.component);
    },
    addTSType(state, v) {
      Vue.set(state.types, v.key, v);
    }
  },
};
