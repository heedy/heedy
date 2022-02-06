<template >
  <h-card-page title="User Settings">
    <v-tabs v-model="tab" show-arrows>
      <v-tabs-slider color="teal lighten-3"></v-tabs-slider>
      <v-tab v-for="r in routes" :key="r.path" :to="`/settings/${r.path}`">{{
        r.title !== undefined ? r.title : r.path
      }}</v-tab>
    </v-tabs>
    <router-view></router-view>
  </h-card-page>
</template>
<script>
export default {
  data: () => ({
    tab: null,
  }),
  computed: {
    routes() {
      return this.$store.state.heedy.settings_routes;
    },
  },
};
</script>