<template>
    <v-app id="heedy" v-resize="onResize">
        <v-navigation-drawer
            :expand-on-hover="true"
            stateless
            app
            dark
            hide-overlay
            :value="!bottom"
            v-if="shownav && !bottom"
        >
            <v-layout column fill-height>
                <v-list dense nav class="py-0">
                    <v-list-item v-if="user==null" two-line to="/login">
                        <v-list-item-avatar>
                            <v-icon>fas fa-sign-in-alt</v-icon>
                        </v-list-item-avatar>

                        <v-list-item-content>
                            <v-list-item-title>Log In</v-list-item-title>
                            <v-list-item-subtitle>Access your Heedy account</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>
                    <v-list-item v-else two-line  :to="'/users/' + user.username">
                        <v-list-item-avatar>
                            <avatar :image="user.avatar" :colorHash="user.username"></avatar>
                        </v-list-item-avatar>

                        <v-list-item-content>
                            <v-list-item-title>{{ username }}</v-list-item-title>
                            <v-list-item-subtitle>{{ user.name }}</v-list-item-subtitle>
                        </v-list-item-content>
                    </v-list-item>

                    <v-divider></v-divider>

                    <v-list-item v-for="item in menu" :key="item.key" :to="item.route">
                        <v-list-item-avatar>
                            <v-icon>{{ item.icon }}</v-icon>
                        </v-list-item-avatar>

                        <v-list-item-content>
                            <v-list-item-title>{{ item.text }}</v-list-item-title>
                        </v-list-item-content>
                    </v-list-item>

                    <v-menu right v-if="showSecondaryMenu" >
                        <template #activator="{ on }">
                            <v-list-item v-on="on">
                                <v-list-item-avatar>
                                    <v-icon>more_vert</v-icon>
                                </v-list-item-avatar>

                                <v-list-item-content>
                                    <v-list-item-title>More</v-list-item-title>
                                </v-list-item-content>
                            </v-list-item>
                        </template>
                        <v-list>
                            <v-list-item v-for="item in secondaryMenu" :key="item.key" :to="item.route">
                                <v-list-item-avatar>
                                    <v-icon>{{ item.icon }}</v-icon>
                                </v-list-item-avatar>
                                <v-list-item-title>{{ item.text }}</v-list-item-title>
                            </v-list-item>
                            <v-list-item v-if="user!=null" to="/logout">
                                <v-list-item-avatar>
                                    <v-icon>fas fa-sign-out-alt</v-icon>
                                </v-list-item-avatar>
                                <v-list-item-title>Log Out</v-list-item-title>
                            </v-list-item>
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
            <v-btn dark text @click="alert_value = false">Close</v-btn>
        </v-snackbar>

        <router-view></router-view>
        <v-bottom-navigation
            dark
            v-if="bottom && shownav"
            app
            grow
            >
            <v-btn v-if="user==null" to="/login">
                <span>Log In</span>
                <v-icon>fas fa-sign-in-alt</v-icon>
            </v-btn>
            <v-btn v-else :to="'/users/' + user.username">
                <span>{{ username }}</span>
                <avatar :image="user.avatar" :colorHash="user.username" :size="28"></avatar>
            </v-btn>

            <v-btn v-for="item in menu" :key="item.key" :to="item.route">
                <span>{{ item.text }}</span>
                <v-icon>{{ item.icon }}</v-icon>
            </v-btn>

            <v-menu offset-y top v-if="showSecondaryMenu">
                <template #activator="{ on }">
                    <v-btn v-on="on">
                        <span>More</span>
                        <v-icon>more_vert</v-icon>
                    </v-btn>
                </template>
                <v-list>
                    <v-list-item v-for="item in secondaryMenu" :key="item.key" :to="item.route">
                        <v-list-item-avatar>
                            <v-icon>{{ item.icon }}</v-icon>
                        </v-list-item-avatar>
                        <v-list-item-title>{{ item.text }}</v-list-item-title>
                    </v-list-item>
                    <v-list-item v-if="user!=null" to="/logout">
                        <v-list-item-avatar>
                            <v-icon>fas fa-sign-out-alt</v-icon>
                        </v-list-item-avatar>
                        <v-list-item-title>Log Out</v-list-item-title>
                    </v-list-item>
                </v-list>
            </v-menu>
        </v-bottom-navigation>
    </v-app>
</template>
<script>

import {Avatar} from "../components.mjs";

export default {
    components: {
        Avatar
    },
    data: () => ({
        bottom: false, // Whether to display the navigation on bottom, in mobile mode
        small: false, // In mobile mode whether to show text. Only active when mini is true
    }),
    computed: {
        menu() {
        let s = this.$store.state.app.menu;
        return s;
        },
        secondaryMenu() {
          return this.$store.state.app.secondaryMenu;
        },
        user() {
        return this.$store.state.app.info.user;
        },
        shownav() {
          return true;
          //return Object.keys(this.$store.state.app.menu).length > 0; // Only show the nav if there is a menu to show.
        },
        showSecondaryMenu() {
          // Must have logout button
          if (this.$store.state.app.info.user!=null) {
            return true;
          }
          // Otherwise, only show it if there are menu items for it
          return this.$store.state.app.secondaryMenu.length > 0;
        },
        username() {
        let u = this.$store.state.app.info.user;
        if (u.name.length == 0) {
            return u.username;
        }
        return u.name.length > 10 ? u.name.split(" ")[0] : u.name;
        },
        alert() {
        return this.$store.state.heedy.alert;
        },
        alert_value: {
        get() {
            return this.$store.state.heedy.alert.value;
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
    }
  }
}
</script>