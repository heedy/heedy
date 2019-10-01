<template>
  <v-badge overlap color="orange" :value="notifications>0">
    <template v-slot:badge>{{ notifications }}</template>
    <v-icon>notifications</v-icon>
  </v-badge>
</template>
<script>
export default {
  computed: {
    notifications() {
      let gn = this.$store.state.notifications.global;
      if (gn == null) {
        return 0;
      }
      return Object.values(gn).reduce((i, n) => (n.seen ? i : i + 1), 0);
    }
  },
  created() {
    this.$store.dispatch("readGlobalNotifications");
  }
};
</script>