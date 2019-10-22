import Vue from "../../dist/vue.mjs";
// import api from "../../api.mjs";


export default {
    state: {
        visualizations: {},
        datavis: {}
    },
    mutations: {
        addVisualization(state, v) {
            Vue.set(state.visualizations, v.key, v.component);
        },
        setData(state, v) {
            console.log("setdata", v);
            Vue.set(state.datavis, v.id, v.data);
        }
    }
};