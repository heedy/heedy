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
      <v-toolbar flat class="transparent">
        <v-list class="pa-0">
          <v-list-tile avatar>
            <v-list-tile-avatar>
              <img src="https://randomuser.me/api/portraits/men/85.jpg">
            </v-list-tile-avatar>

            <v-list-tile-content>
              <v-list-tile-title>Daniel Kumor</v-list-tile-title>
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

        <v-list-tile v-for="item in items" :key="item.title">
          <v-list-tile-action>
            <v-icon>{{ item.icon }}</v-icon>
          </v-list-tile-action>

          <v-list-tile-content>
            <v-list-tile-title>{{ item.title }}</v-list-tile-title>
          </v-list-tile-content>
        </v-list-tile>
      </v-list>
    </v-navigation-drawer>

    <user/>
  </v-app>
</template>

<script>
export default {
  data: () => ({
    drawer: true,
    items: [
      { title: "Home", icon: "dashboard" },
      { title: "About", icon: "question_answer" }
    ],
    mini: null
  }),
  props: {
    source: String
  }
};
</script>

<style>
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