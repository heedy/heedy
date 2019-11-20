<template>
  <v-flex>
    <h-object-list :objects="objects" />
  </v-flex>
</template>
<script>
export default {
  props: {
    app: Object
  },
  computed: {
    objects() {
      return Object.keys(
        this.$store.state.heedy.appObjects[this.app.id] || {}
      ).map(id => this.$store.state.heedy.objects[id]);
    }
  },
  watch: {
    app: function(c) {
      this.$store.dispatch("readAppObjects", { id: c.id });
    }
  },

  created() {
    this.$store.dispatch("readAppObjects", { id: this.app.id });
  }
};
</script>