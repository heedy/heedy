<template>
  <h-page-container>
    <v-card>
      <v-card-title>
        <v-list-item>
          <v-list-item-content>
            <v-list-item-title class="headline mb-1">Notifications</v-list-item-title>
          </v-list-item-content>
        </v-list-item>
      </v-card-title>
      <v-container>
        <div v-if="loading" style="color: gray; text-align: center;">Loading...</div>
        <div
          v-else-if="notifications.length==0"
          style="color: gray; text-align: center;"
        >You don't have any notifications.</div>
        <v-row no-gutters v-else>Hi</v-row>
      </v-container>
    </v-card>
  </h-page-container>
</template>
<script>
export default {
  computed: {
    loading() {
      return this.$store.state.notifications.global == null;
    },
    notifications() {
      return this.$store.state.notifications.global;
    }
  },
  created() {
    this.$store.dispatch("readGlobalNotifications");
  }
};
</script>