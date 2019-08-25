import Vue from "vue";
import Vuetify from "vuetify";
import VueRouter from "vue-router";
import Vuex, { mapState } from "vuex";
import VueHeadful from "vue-headful";

import VueCodemirror from 'vue-codemirror';

import 'codemirror/mode/javascript/javascript.js';
import 'codemirror/mode/python/python.js';
import 'codemirror/lib/codemirror.css';

// For some reason, postcss refuses to load the
// vuetify css. We therefore manually include it
// in the html.
// import 'vuetify/dist/vuetify.min.css';
//import '@mdi/font/css/materialdesignicons.css';
// require styles


// Disable the vue console messages
Vue.config.productionTip = false;
Vue.config.devtools = false;

Vue.use(Vuetify);
Vue.use(VueRouter);
Vue.use(Vuex);
Vue.use(VueCodemirror);

// Setting the title component
Vue.component('vue-headful', VueHeadful);


// Export the libraries
export {
    VueRouter,Vuex,Vuetify,VueCodemirror,mapState
};

export default Vue;