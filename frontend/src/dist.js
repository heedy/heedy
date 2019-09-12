import Vue from "vue";
import VueRouter from "vue-router";
import Vuex, { mapState } from "vuex";
import VueHeadful from "vue-headful";

import Vuetify from "vuetify";
import 'vuetify/dist/vuetify.min.css';

// We use python and javascript (json) for codemirror
import VueCodemirror from 'vue-codemirror';
import 'codemirror/lib/codemirror.css';
import 'codemirror/mode/javascript/javascript.js';
import 'codemirror/mode/python/python.js';

// Include all icon libraries
import '@fortawesome/fontawesome-free/css/all.css';
//import "@fortawesome/fontawesome-free/webfonts/fa-regular-400.woff2";

import 'material-design-icons-iconfont/dist/material-design-icons.css';

// Include the roboto font
import "typeface-roboto";

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