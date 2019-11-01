<template>
  <v-flex v-if="hasNotifications">
    <h-notification
      v-for="n in notifications"
      :key="n.key+'.'+n.user + '.' + n.app + '.' + n.source"
      :n="n"
      small
      seen
    />
  </v-flex>
</template>
<script>
export default {
  props: {
    app: Object
  },
  computed: {
    hasNotifications() {
      let narr = this.notifications;
      return narr != null && narr.length > 0;
    },
    notifications() {
      let n = this.$store.state.notifications.apps[this.app.id] || null;
      if (n == null) return null;
      let narr = Object.values(n);
      narr.sort((a, b) => b.timestamp - a.timestamp);
      return narr;
    }
  },
  watch: {
    app: function(newValue) {
      this.$store.dispatch("readAppNotifications", { id: newValue.id });
    }
  },
  created() {
    this.$store.dispatch("readAppNotifications", {
      id: this.app.id
    });
  }
};
</script>