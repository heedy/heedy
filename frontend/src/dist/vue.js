import Vue from "vue";
import VueRouter from "vue-router";
import Vuex, { mapState } from "vuex";
import createLogger from "vuex/dist/logger";

// Vuetify internally imports vue, so we need to include it here.
import Vuetify from "vuetify";
import "vuetify/dist/vuetify.min.css";
import "typeface-roboto";

// Fontawesome is used from vuetify
import "@fortawesome/fontawesome-free/css/all.css";
import "material-design-icons-iconfont/dist/material-design-icons.css";

// Add Vuetify-jsonschema-form
import VJsf from "@koumoul/vjsf";
import "@koumoul/vjsf/dist/main.css";

// Disable the vue console messages if built with production
Vue.config.productionTip = false;

Vue.component("VJsf", VJsf.default);

Vue.use(VueRouter);
Vue.use(Vuex);
Vue.use(Vuetify);

export { VJsf, VueRouter, Vuex, Vuetify, mapState, createLogger };
export default Vue;
