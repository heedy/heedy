import Vue from "../../dist/vue.mjs";
// import api from "../../api.mjs";


export default {
    state: {
        views: {},
        streams: {}
    },
    mutations: {
        addView(state, v) {
            Vue.set(state.views, v.key, v.component);
        },
        setData(state, v) {
            Vue.set(state.streams, v.id, v.data);
        }
    }
};