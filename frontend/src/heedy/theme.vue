<template>
  <v-app id="heedy" v-resize="onResize">
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
      <v-layout column fill-height>
        <v-toolbar flat class="transparent">
          <v-list class="pa-0" v-if="user==null">
            <v-tooltip right dark :disabled="!mini">
              <v-list-tile
                avatar
                to="/login"
                slot="activator"
                active-class="active-btn"
                class="inactive-btn"
              >
                <v-list-tile-avatar>
                  <v-icon>fas fa-sign-in-alt</v-icon>
                </v-list-tile-avatar>
                <v-list-tile-content>
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
                  <avatar :image="user.avatar"></avatar>
                </v-list-tile-avatar>

                <v-list-tile-content>
                  <v-list-tile-title>{{ username }}</v-list-tile-title>
                </v-list-tile-content>
              </v-list-tile>
              <span>{{ username }}</span>
            </v-tooltip>
          </v-list>
        </v-toolbar>
        <v-list class="pt-0" dense>
          <v-divider v-if="!transparent"></v-divider>
          <v-tooltip v-for="item in menu" :key="item.key" dark right :disabled="!mini">
            <v-list-tile
              :to="item.route"
              active-class="active-btn"
              class="inactive-btn"
              slot="activator"
              avatar
            >
              <v-list-tile-avatar>
                <v-icon
                  v-if="item.icon.startsWith('fa:') || item.icon.startsWith('mi:')"
                >{{ item.icon.substring(3,item.icon.length) }}</v-icon>
                <img v-else :src="item.icon">
              </v-list-tile-avatar>

              <v-list-tile-content>
                <v-list-tile-title>{{ item.text }}</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
            <span>{{ item.text }}</span>
          </v-tooltip>
        </v-list>
        <v-spacer></v-spacer>

        <v-list class="pt-0" dense v-if="user!=null">
          <!-- https://github.com/vuetifyjs/vuetify/issues/4848 -->
          <v-menu offset-x style="width:100%">
            <template #activator="{ on: menu }">
              <v-tooltip right dark :disabled="!mini">
                <template #activator="{ on: tooltip }">
                  <v-list-tile slot="activator" v-on="{ ...tooltip, ...menu }" style="width:100%">
                    <v-list-tile-avatar>
                      <v-icon>more_vert</v-icon>
                    </v-list-tile-avatar>

                    <v-list-tile-content v-if="!mini">
                      <v-list-tile-title>More</v-list-tile-title>
                    </v-list-tile-content>
                  </v-list-tile>
                </template>
                <span>More</span>
              </v-tooltip>
            </template>
            <v-list>
              <v-list-tile avatar to="/settings">
                <v-list-tile-avatar>
                  <v-icon>settings</v-icon>
                </v-list-tile-avatar>
                <v-list-tile-content>
                  <v-list-tile-title>Settings</v-list-tile-title>
                </v-list-tile-content>
              </v-list-tile>
              <v-list-tile avatar to="/logout">
                <v-list-tile-avatar>
                  <v-icon>fas fa-sign-out-alt</v-icon>
                </v-list-tile-avatar>
                <v-list-tile-content>
                  <v-list-tile-title>Log Out</v-list-tile-title>
                </v-list-tile-content>
              </v-list-tile>
            </v-list>
          </v-menu>
        </v-list>
      </v-layout>
    </v-navigation-drawer>

    <v-snackbar
      v-model="alert_value"
      :color="alert.type"
      :timeout="4000"
      :vertical="false"
      top
      :right="!bottom"
    >
      {{ alert.text }}
      <v-btn dark flat @click="alert_value = false">Close</v-btn>
    </v-snackbar>

    <router-view></router-view>

    <v-bottom-nav
      :dark="dark"
      :value="bottom"
      app
      :class="{'theme-dark': !transparent, 'transparent': transparent,'clearshadows': transparent}"
      v-if="shownav"
    >
      <v-tooltip top dark :disabled="!small" v-if="user==null" style="text-align:center;">
        <v-btn dark flat to="/login" slot="activator">
          <span v-if="!small">Log In</span>
          <v-icon>fas fa-sign-in-alt</v-icon>
        </v-btn>
        <span>Log In</span>
      </v-tooltip>
      <v-tooltip v-else top dark :disabled="!small" style="text-align:center;">
        <v-btn dark flat :to="'/user/' + user.name" slot="activator">
          <span v-if="!small" style="padding-top: 5px;">{{ username }}</span>
          <avatar :image="user.avatar" :size="28"></avatar>
        </v-btn>
        <span>{{ username }}</span>
      </v-tooltip>
      <v-tooltip
        v-for="item in menu"
        :key="item.key"
        top
        dark
        :disabled="!small"
        style="text-align:center;"
      >
        <v-btn dark flat :to="item.route" slot="activator">
          <span v-if="!small">{{ item.text }}</span>
          <v-icon
            v-if="item.icon.startsWith('fa:') || item.icon.startsWith('mi:')"
          >{{ item.icon.substring(3,item.icon.length) }}</v-icon>
          <img v-else :src="item.icon">
        </v-btn>
        <span>{{ item.text }}</span>
      </v-tooltip>
      <div style="text-align:center;" v-if="user!=null">
        <v-menu offset-y top>
          <template #activator="{ on: menu }">
            <v-tooltip top dark :disabled="!small">
              <template #activator="{ on: tooltip }">
                <v-btn dark flat v-on="{ ...tooltip, ...menu }">
                  <span v-if="!small">More</span>
                  <v-icon>more_vert</v-icon>
                </v-btn>
              </template>
              <span>More</span>
            </v-tooltip>
          </template>
          <v-list>
            <v-list-tile avatar to="/settings">
              <v-list-tile-avatar>
                <v-icon>settings</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>Settings</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
            <v-list-tile avatar to="/logout">
              <v-list-tile-avatar>
                <v-icon>fas fa-sign-out-alt</v-icon>
              </v-list-tile-avatar>
              <v-list-tile-content>
                <v-list-tile-title>Log Out</v-list-tile-title>
              </v-list-tile-content>
            </v-list-tile>
          </v-list>
        </v-menu>
      </div>
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

import Avatar from "./components/avatar.mjs";

export default {
  components: {
    Avatar
  },
  data: () => ({
    bottom: false, // Whether to display the navigation on bottom, in mobile mode
    mini: true, // In desktop mode, whether to show mini drawer
    small: false, // In mobile mode whether to show text. Same effect as mini
    transparent: false, // Whether the nav is to be transparent to fit in with the page theme
    dark: true // Whether to user the dark theme
  }),
  props: {
    source: String
  },
  computed: {
    menu() {
      let s = this.$store.state.info.menu;
      return Object.keys(s).map(k => ({
        key: k,
        text: s[k].text,
        icon: s[k].icon,
        route: s[k].route
      }));
    },
    user() {
      return this.$store.state.info.user;
    },
    shownav() {
      return Object.keys(this.$store.state.info.menu).length > 0; // Only show the nav if there is a menu to show.
    },
    username() {
      let u = this.$store.state.info.user;
      if (u.fullname.length == 0) {
        return u.name;
      }
      return u.fullname.length > 10 ? u.fullname.split(" ")[0] : u.fullname;
    },
    alert() {
      return this.$store.state.alert;
    },
    alert_value: {
      get() {
        return this.$store.state.alert.value;
      },
      set(newValue) {
        this.$store.commit("alert", {
          value: newValue,
          text: "",
          type: "info"
        });
      }
    }
  },
  mounted() {
    this.onResize();
  },

  methods: {
    onResize() {
      this.bottom = window.innerWidth < 960;
      this.small = window.innerWidth < 500;
      this.mini = window.innerWidth < 1440;
    }
  }
};
</script>

<style>
.active-btn {
  color: white !important;
}

.inactive-btn {
  color: #cccccc;
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
