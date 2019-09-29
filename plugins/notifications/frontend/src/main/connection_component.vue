<template>
  <v-flex v-if="hasNotifications">
    <h-notification
      v-for="n in notifications"
      :key="n.key+'.'+n.user + '.' + n.connection + '.' + n.source"
      :n="n"
      small
    />
  </v-flex>
</template>
<script>
export default {
  props: {
    connection: Object
  },
  computed: {
    hasNotifications() {
      let narr = this.notifications;
      return narr != null && narr.length > 0;
    },
    notifications() {
      let n =
        this.$store.state.notifications.connections[this.connection.id] || null;
      if (n == null) return null;
      let narr = Object.values(n);
      narr.sort((a, b) => b.timestamp - a.timestamp);
      return narr;
    }
  },
  watch: {
    connection: function(newValue) {
      this.$store.dispatch("readConnectionNotifications", { id: newValue.id });
    }
  },
  created() {
    this.$store.dispatch("readConnectionNotifications", {
      id: this.connection.id
    });
  }
};
</script>