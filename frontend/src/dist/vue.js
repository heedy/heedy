import Vue from "vue";
import VueCompositionAPI, { ref, reactive } from '@vue/composition-api'
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
import "regenerator-runtime/runtime"; // Needed to vjsf file upload support https://github.com/koumoul-dev/vuetify-jsonschema-form/issues/301

// Disable the vue console messages if built with production
Vue.config.productionTip = false;

Vue.use(VueCompositionAPI);

Vue.component("VJsf", VJsf);

Vue.use(VueRouter);
Vue.use(Vuex);
Vue.use(Vuetify);

export { ref,reactive, VJsf, VueRouter, Vuex, Vuetify, mapState, createLogger };
export default Vue;
