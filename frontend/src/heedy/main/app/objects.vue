<template>
  <v-flex>
    <h-object-list :objects="objects" :showApps="false" />
  </v-flex>
</template>
<script>
export default {
  props: {
    app: Object,
  },
  computed: {
    objects() {
      return Object.keys(
        this.$store.state.heedy.appObjects[this.app.id] || {}
      ).map((id) => this.$store.state.heedy.objects[id]);
    },
    websocket() {
      return this.$store.state.app.websocket!=null;
    }
  },
  watch: {
    app: function (c,oc) {
      if (c.id!=oc.id) {
        this.$store.dispatch("readAppObjects", { id: c.id });
      }
    },
    websocket(nv) {
      if (nv) {
        // If the websocket gets re-connected, re-read the objects
        this.$store.dispatch("readAppObjects", {
          id: this.app.id
        });
      }
    }
  },

  created() {
    this.$store.dispatch("readAppObjects", { id: this.app.id });
  },
};
</script>