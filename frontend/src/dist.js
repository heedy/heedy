import Vue from "vue";
import VueRouter from "vue-router";
import Vuex, {
    mapState
} from "vuex";
import VueHeadful from "vue-headful";

import Vuetify from "vuetify";
import 'vuetify/dist/vuetify.min.css';
import "typeface-roboto";

import VueCodemirror from 'vue-codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/javascript/javascript.js';
import 'codemirror/mode/python/python.js';

import '@fortawesome/fontawesome-free/css/all.css';
import 'material-design-icons-iconfont/dist/material-design-icons.css';

import moment from "moment";

// Disable the vue console messages
Vue.config.productionTip = false;
Vue.config.devtools = false;

Vue.use(Vuetify);
Vue.use(VueRouter);
Vue.use(Vuex);
Vue.use(VueCodemirror);

// Setting the title component
Vue.component('vue-headful', VueHeadful);

export {
    VueRouter,
    Vuex,
    Vuetify,
    VueCodemirror,
    mapState,
    moment
};

export default Vue;