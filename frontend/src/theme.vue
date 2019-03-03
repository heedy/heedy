<template>
  <v-app id="inspire">
    <v-navigation-drawer
      v-model="drawer"
      :mini-variant.sync="mini"
      hide-overlay
      stateless
      aria-haspopup="clipped"
      app
      class="theme-dark"
      dark
    >
      <v-toolbar flat class="transparent" v-if="user!=null">
        <v-list class="pa-0">
          <v-list-tile avatar :to="'/user/' + user.name">
            <v-list-tile-avatar>
              <v-icon v-if="user.icon.startsWith('fa:') || user.icon.startsWith('mi:')">{{ user.icon.substring(3,user.icon.length) }}</v-icon>
              <img v-else :src="user.icon" />
            </v-list-tile-avatar>

            <v-list-tile-content>
              <v-list-tile-title>{{ user.fullname }}</v-list-tile-title>
            </v-list-tile-content>

            <v-list-tile-action>
              <v-btn icon @click.stop="mini = !mini">
                <v-icon>chevron_left</v-icon>
              </v-btn>
            </v-list-tile-action>
          </v-list-tile>
        </v-list>
        
      </v-toolbar>
      
      <v-list class="pt-0" dense>
        <v-divider></v-divider>
        <v-list-tile v-for="item in menu" :key="item.key" :to="item.route" active-class="active-btn">
          <v-list-tile-action>
            <v-icon v-if="item.icon.startsWith('fa:') || item.icon.startsWith('mi:')">{{ item.icon.substring(3,item.icon.length) }}</v-icon>
            <img v-else :src="item.icon" />
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>{{ item.text }}</v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>
      </v-list>
    </v-navigation-drawer>

    <router-view></router-view>
  </v-app>
</template>

<script>
/*
import Vue from "vue";
import Vuetify from "vuetify";
import "vuetify/dist/vuetify.min.css";
*/

// Vue is used as a global
//Vue.use(Vuetify);

export default {
  data: () => ({
    drawer: true,
    items: [
      { title: "Home", icon: "dashboard" },
      { title: "About", icon: "question_answer" }
    ],
    mini: true
  }),
  props: {
    source: String
  },
  computed: {
    menu() {
      let s = this.$store.state.menu;
      return Object.keys(s).map(k => ({key: k, text: s[k].text, icon: s[k].icon, route: s[k].route}));
    },
    user() {
      return {fullname: "Daniel Kumor",name: "dkumor", icon:"mi:face"};//this.$store.state.user;
    }
  }
};
</script>

<style>
.active-btn {
  color: #DCDCDC;
}
.theme-dark {
  color: #1c313a;
  background-color: #1c313a !important;
}
.theme-primary {
  color: #455a64;
  background-color: #455a64 !important;
}
.theme-light {
  color: #718792;
  background-color: #455a64 !important;
}
</style>