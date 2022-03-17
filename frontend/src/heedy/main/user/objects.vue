<template>
  <v-flex>
    <h-object-list :objects="objects" />
  </v-flex>
</template>
<script>
export default {
  props: {
    user: Object
  },
  computed: {
    objects() {
      return Object.keys(
        this.$store.state.heedy.userObjects[this.user.username] || {}
      ).map(id => this.$store.state.heedy.objects[id]);
    },
    websocket() {
      return this.$store.state.app.websocket!=null;
    }
  },
  watch: {
    user: function(u,ou) {
      if (u.username!=ou.username) {
        this.$store.dispatch("readUserObjects", { username: u.username });
      }
    },
    websocket(nv) {
      if (nv) {
        // If the websocket gets re-connected, re-read the objects
        this.$store.dispatch("readUserObjects", {
          username: this.user.username
        });
      }
    }
  },

  created() {
    this.$store.dispatch("readUserObjects", { username: this.user.username });
  }
};
</script>