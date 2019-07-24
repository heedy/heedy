import Vue from "vue";
import Vuetify from "vuetify";
import VueRouter from "vue-router";
import Vuex, { mapState } from "vuex";

// For some reason, postcss refuses to load the
// vuetify css. We therefore manually include it
// in the html.
// import 'vuetify/dist/vuetify.min.css';
import '@mdi/font/css/materialdesignicons.css';

Vue.config.productionTip = false;
Vue.use(Vuetify);
Vue.use(VueRouter);
Vue.use(Vuex);

// Export the libraries
export {
    VueRouter,Vuex,Vuetify,mapState
};

export default Vue;