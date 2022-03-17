<template>
  <v-flex v-if="hasNotifications">
    <h-notification
      v-for="n in notifications"
      :key="n.key+'.'+n.user + '.' + n.app + '.' + n.object"
      :n="n"
      small
    />
  </v-flex>
</template>
<script>
export default {
  props: {
    object: Object
  },
  computed: {
    hasNotifications() {
      let narr = this.notifications;
      return narr != null && narr.length > 0;
    },
    notifications() {
      let n = this.$store.state.notifications.objects[this.object.id] || null;
      if (n == null) return null;
      let narr = Object.values(n);
      narr.sort((a, b) => b.timestamp - a.timestamp);
      return narr;
    },
    websocket() {
      return this.$store.state.app.websocket!=null;
    }
  },
  watch: {
    object: function(newValue,oldValue) {
      if (newValue.id!=oldValue.id) {
        this.$store.dispatch("readObjectNotifications", { id: newValue.id });
      }
    },
    websocket(nv) {
      if (nv) {
        // If the websocket gets re-connected, re-read notifications
        this.$store.dispatch("readObjectNotifications", {
          id: this.object.id
        });
      }
    }
  },
  created() {
    this.$store.dispatch("readObjectNotifications", {
      id: this.object.id
    });
  }
};
</script>