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
    source: Object
  },
  computed: {
    hasNotifications() {
      let narr = this.notifications;
      return narr != null && narr.length > 0;
    },
    notifications() {
      let n = this.$store.state.notifications.sources[this.source.id] || null;
      if (n == null) return null;
      let narr = Object.values(n);
      narr.sort((a, b) => b.timestamp - a.timestamp);
      return narr;
    }
  },
  watch: {
    source: function(newValue) {
      this.$store.dispatch("readSourceNotifications", { id: newValue.id });
    }
  },
  created() {
    this.$store.dispatch("readSourceNotifications", {
      id: this.source.id
    });
  }
};
</script>