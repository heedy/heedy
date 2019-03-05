<template>
  <v-app id="cdb" v-resize="onResize">
    <v-navigation-drawer
      :mini-variant="mini"
      stateless
      :value="!bottom"
      :class="{'theme-dark': !transparent, 'transparent': transparent,'clearshadows': transparent}"
      width="200"
      app
      :floating="transparent"
      :dark="dark"
      hide-overlay
      :temporary="transparent"
      v-if="shownav"
    >
      <v-toolbar flat class="transparent" >
        <v-list class="pa-0" v-if="user==null">
          <v-tooltip right dark :disabled="!mini">
              <v-list-tile avatar to="/login" slot="activator">
                <v-list-tile-avatar>
                  <v-icon>fas fa-sign-in-alt</v-icon>
                </v-list-tile-avatar>
                <v-list-tile-content >
                  <v-list-tile-title>Log In</v-list-tile-title>
                </v-list-tile-content>
              </v-list-tile>
            <span>Log In</span>
          </v-tooltip>
        </v-list>
        <v-list class="pa-0" v-else>
          <v-tooltip right dark :disabled="!mini">
            <v-list-tile avatar :to="'/user/' + user.name" slot="activator">
              <v-list-tile-avatar>
                <v-icon v-if="user.icon.startsWith('fa:') || user.icon.startsWith('mi:')">{{ user.icon.substring(3,user.icon.length) }}</v-icon>
                <img v-else :src="user.icon" />
              </v-list-tile-avatar>

              <v-list-tile-content >
                <v-list-tile-title>{{ user.fullname }}</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
            <span>{{ user.fullname }}</span>
          </v-tooltip>
        </v-list>
      </v-toolbar>
      
      <v-list class="pt-0" dense>
        <v-divider v-if="!transparent"></v-divider>
        <v-tooltip v-for="item in menu" :key="item.key" dark :disabled="!mini" right>
          <v-list-tile :to="item.route" active-class="active-btn" slot="activator">
            <v-list-tile-action>
              <v-icon v-if="item.icon.startsWith('fa:') || item.icon.startsWith('mi:')">{{ item.icon.substring(3,item.icon.length) }}</v-icon>
              <img v-else :src="item.icon" />
            </v-list-tile-action>

            <v-list-tile-content>
              <v-list-tile-title>{{ item.text }}</v-list-tile-title>
            </v-list-tile-content>
          </v-list-tile>
          <span>{{ item.text }}</span>
        </v-tooltip>
      </v-list>
    </v-navigation-drawer>
      
    <router-view></router-view>
    
    <v-bottom-nav
        :dark="dark"
        :value="bottom"
        app
        :class="{'theme-dark': !transparent, 'transparent': transparent,'clearshadows': transparent}"
        v-if="shownav"
      >
      <v-tooltip top dark :disabled="!small" v-if="user==null">
        <v-btn dark flat  to="/login" slot="activator">
          <span v-if="!small">Log In</span>
          <v-icon>fas fa-sign-in-alt</v-icon>
        </v-btn>
        <span>Log In</span>
      </v-tooltip>
      <v-tooltip v-else top dark :disabled="!small">
        <v-btn dark flat  :to="'/user/' + user.name" slot="activator">
          <span v-if="!small">{{ user.fullname }}</span>
          <v-icon v-if="user.icon.startsWith('fa:') || user.icon.startsWith('mi:')">{{ user.icon.substring(3,user.icon.length) }}</v-icon>
          <img v-else :src="user.icon" />
        </v-btn>
        <span>{{ user.fullname }}</span>
      </v-tooltip>
      <v-tooltip v-for="item in menu" :key="item.key" top dark :disabled="!small" >
        <v-btn dark flat  :to="item.route" slot="activator">
          <span v-if="!small">{{ item.text }}</span>
          <v-icon v-if="item.icon.startsWith('fa:') || item.icon.startsWith('mi:')">{{ item.icon.substring(3,item.icon.length) }}</v-icon>
          <img v-else :src="item.icon" />
        </v-btn>
        <span>{{ item.text }}</span>
      </v-tooltip>
    </v-bottom-nav>
  </v-app>
</template>

<script>

/*
import Vue from "vue";
import Vuetify from "vuetify";
import "vuetify/dist/vuetify.min.css";

// This theme uses Vuetify. All pages are free to use vuetify components.
// This also means that any theme must either include vuetify, or replace/reimplement all frontend pages.
Vue.use(Vuetify);
*/

export default {
  data: () => ({
    bottom: false,  // Whether to display the navigation on bottom, in mobile mode
    mini: true,      // In desktop mode, whether to show mini drawer
    small: false,    // In mobile mode whether to show text. Same effect as mini
    transparent: false, // Whether the nav is to be transparent to fit in with the page theme
    dark: true  // Whether to user the dark theme
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
      return this.$store.state.user;
    },
    shownav() {
      return Object.keys(this.$store.state.menu).length > 0; // Only show the nav if there is a menu to show.
    }
  },
  mounted () {
      this.onResize()
    },

    methods: {
      onResize () {
        this.bottom = (window.innerWidth < 960);
        this.small = (window.innerWidth < 500);
        this.mini = (window.innerWidth < 1440);
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

.clearshadows {
  box-shadow: none !important;
}
</style>