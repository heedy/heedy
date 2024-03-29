<template>
  <v-app id="heedy" v-resize="debounceResize">
    <v-navigation-drawer
      expand-on-hover
      permanent
      stateless
      app
      dark
      hide-overlay
      :mini-variant-width="70"
      v-if="shownav && !bottom"
    >
      <v-layout column fill-height>
        <v-list dense nav>
          <v-list-item v-if="user == null" two-line to="/login">
            <v-list-item-avatar>
              <v-icon>fas fa-sign-in-alt</v-icon>
            </v-list-item-avatar>

            <v-list-item-content>
              <v-list-item-title>Log In</v-list-item-title>
              <v-list-item-subtitle
                >Access your Heedy account</v-list-item-subtitle
              >
            </v-list-item-content>
          </v-list-item>
          <v-list-item v-else two-line :to="'/users/' + user.username">
            <v-list-item-avatar>
              <h-icon
                :image="user.icon"
                defaultIcon="person"
                :colorHash="user.username"
              ></h-icon>
            </v-list-item-avatar>

            <v-list-item-content>
              <v-list-item-title>{{ username }}</v-list-item-title>
              <v-list-item-subtitle>{{ user.username }}</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>

          <v-divider></v-divider>

          <v-list-item
            v-for="item in menu.primary"
            :key="item.key"
            :to="item.route"
          >
            <v-list-item-avatar>
              <v-icon v-if="item.component === undefined">{{
                item.icon
              }}</v-icon>
              <component
                v-else
                :is="item.component"
                :status="bottom ? 'bottom' : 'side'"
              />
            </v-list-item-avatar>

            <v-list-item-content>
              <v-list-item-title>{{ item.text }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
        </v-list>
        <v-spacer></v-spacer>
        <v-list dense nav class="py-0">
          <v-list-item
            v-for="item in menu.bottom"
            :key="item.key"
            :to="item.route"
          >
            <v-list-item-icon>
              <v-icon v-if="item.component === undefined">{{
                item.icon
              }}</v-icon>
              <component
                v-else
                :is="item.component"
                :status="bottom ? 'bottom' : 'side'"
              />
            </v-list-item-icon>

            <v-list-item-content>
              <v-list-item-title>{{ item.text }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>

          <v-menu right v-if="menu.showSecondary">
            <template #activator="{ on }">
              <v-list-item v-on="on" height="30px">
                <v-list-item-icon>
                  <v-icon>more_vert</v-icon>
                </v-list-item-icon>

                <v-list-item-content>
                  <v-list-item-title>More</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
            </template>
            <v-list
              dense
              nav
              width="200px"
              :style="`max-height: ${height - 50}px`"
              class="overflow-y-auto"
            >
              <v-list-item
                v-for="item in menu.secondary"
                :key="item.key"
                :to="item.route"
              >
                <v-list-item-icon>
                  <v-icon v-if="item.component === undefined">{{
                    item.icon
                  }}</v-icon>
                  <component
                    v-else
                    :is="item.component"
                    :status="bottom ? 'bottom' : 'side'"
                  />
                </v-list-item-icon>
                <v-list-item-title>{{ item.text }}</v-list-item-title>
              </v-list-item>
              <v-list-item v-if="user != null" to="/logout">
                <v-list-item-icon>
                  <v-icon>fas fa-sign-out-alt</v-icon>
                </v-list-item-icon>
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
      <template v-slot:action="{ attrs }">
        <v-btn dark text @click="alert_value = false" v-bind="attrs"
          >Close</v-btn
        >
      </template>
    </v-snackbar>

    <router-view></router-view>
    <v-bottom-navigation dark v-if="bottom && shownav" app grow>
      <v-btn v-if="user == null" to="/login">
        <span v-if="!small">Log In</span>
        <v-icon>fas fa-sign-in-alt</v-icon>
      </v-btn>
      <v-btn v-else :to="'/users/' + user.username">
        <span v-if="!small">{{ username }}</span>
        <h-icon
          :image="user.icon"
          defaultIcon="person"
          :colorHash="user.username"
          :size="28"
        ></h-icon>
      </v-btn>

      <v-btn v-for="item in menu.primary" :key="item.key" :to="item.route">
        <span v-if="!small">{{ item.text }}</span>
        <v-icon v-if="item.component === undefined">{{ item.icon }}</v-icon>
        <component
          v-else
          :is="item.component"
          :status="bottom ? 'bottom' : 'side'"
        />
      </v-btn>
      <v-btn v-for="item in menu.bottom" :key="item.key" :to="item.route">
        <span v-if="!small">{{ item.text }}</span>
        <v-icon v-if="item.component === undefined">{{ item.icon }}</v-icon>
        <component
          v-else
          :is="item.component"
          :status="bottom ? 'bottom' : 'side'"
        />
      </v-btn>

      <v-menu offset-y top v-if="menu.showSecondary">
        <template #activator="{ on }">
          <v-btn v-on="on">
            <span v-if="!small">More</span>
            <v-icon>more_vert</v-icon>
          </v-btn>
        </template>
        <v-list
          dense
          nav
          :style="`max-height: ${height - 50}px`"
          class="overflow-y-auto"
        >
          <v-list-item
            v-for="item in menu.secondary"
            :key="item.key"
            :to="item.route"
          >
            <v-list-item-avatar>
              <v-icon v-if="item.component === undefined">{{
                item.icon
              }}</v-icon>
              <component
                v-else
                :is="item.component"
                :status="bottom ? 'bottom' : 'side'"
              />
            </v-list-item-avatar>
            <v-list-item-title>{{ item.text }}</v-list-item-title>
          </v-list-item>
          <v-list-item v-if="user != null" to="/logout">
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
export default {
  data: () => ({
    bottom: false, // Whether to display the navigation on bottom, in mobile mode
    small: false, // In mobile mode whether to show text. Only active when mini is true
    width: 1000, // The number of pixels available for buttons in the side menu
    height: 1000, // The number of pixels available for buttons in the bottom menu
    resizeTimeout: null, // Debounce timeout for resize event
  }),
  head: {
    title: "heedy",
    titleTemplate: "%s - Heedy",
  },
  computed: {
    menu() {
      let primary = Object.values(this.$store.state.app.menu_items).filter(
        (m) => m.location === undefined || m.location == "primary"
      );
      let bottom = Object.values(this.$store.state.app.menu_items).filter(
        (m) => m.location !== undefined && m.location == "primary_bottom"
      );
      let secondary = Object.values(this.$store.state.app.menu_items).filter(
        (m) =>
          m.location !== undefined &&
          (m.location == "secondary" ||
            (m.location != "primary" && m.location != "primary_bottom"))
      );

      const logoutbtn = this.$store.state.app.info.user != null;

      // Now we have overflow menus for primary and bottom:
      if (this.bottom) {
        const itemSize = this.small ? 80 : 110;
        // In mobile view, the menu is shown on bottom, so is based on width of viewport
        let menuSize = Math.floor((this.width - itemSize * 2) / itemSize);

        if (primary.length > menuSize) {
          bottom = primary.slice(menuSize).concat(bottom);
          primary = primary.slice(0, menuSize);
          menuSize = 0;
        } else {
          menuSize -= primary.length;
        }

        if (bottom.length > menuSize) {
          secondary = bottom.slice(menuSize).concat(secondary);
          bottom = bottom.slice(0, menuSize);
          menuSize = 0;
        } else {
          menuSize -= bottom.length;
        }

        if (secondary.length == 1 && !logoutbtn) {
          // If we don't actually need the overflow menu, so add the button directly
          bottom.push(secondary.pop());
        }
      } else {
        // The side menu has the special top part, and the ... menu
        const mainItemSize = 60;
        let remainingHeight = this.height - 95 - 45;
        let menuSize = Math.floor(remainingHeight / mainItemSize);
        if (primary.length > menuSize) {
          secondary = primary.slice(menuSize).concat(bottom).concat(secondary);
          bottom = [];
          primary = primary.slice(0, menuSize);

          if (secondary.length == 1 && !logoutbtn) {
            // If we don't actually need the overflow menu, so add the button directly
            primary.push(secondary.pop());
          }
        } else {
          remainingHeight -= primary.length * mainItemSize;

          const secondaryItemSize = 43;
          menuSize = Math.floor(remainingHeight / secondaryItemSize);
          if (bottom.length > menuSize) {
            secondary = bottom.slice(menuSize).concat(secondary);
            bottom = bottom.slice(0, menuSize);
          }
          if (secondary.length == 1 && !logoutbtn) {
            // If we don't actually need the overflow menu, so add the button directly
            bottom.push(secondary.pop());
          }
        }
      }

      return { primary, bottom, secondary,showSecondary: (secondary.length > 0 || logoutbtn) };
    },
    user() {
      return this.$store.state.app.info.user;
    },
    shownav() {
      return true;
      //return Object.keys(this.$store.state.app.menu).length > 0; // Only show the nav if there is a menu to show.
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
          type: "info",
        });
      },
    },
  },
  mounted() {
    this.onResize();
  },
  methods: {
    onResize() {
      this.bottom = window.innerWidth < 960;
      this.small = window.innerWidth < 500;

      // The user icon and ... menu are always visible, so remove them from consideration
      this.height = window.innerHeight;
      this.width = window.innerWidth;
    },
    debounceResize() {
      if (this.resizeTimeout != null) {
        clearTimeout(this.resizeTimeout);
      }
      this.resizeTimeout = setTimeout(this.onResize, 100);
    },
    reload() {
      window.location.reload();
    },
  },
};
</script>