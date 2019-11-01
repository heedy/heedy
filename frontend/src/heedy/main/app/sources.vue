<template>
  <v-flex>
    <h-source-list :sources="sources" />
  </v-flex>
</template>
<script>
export default {
  props: {
    app: Object
  },
  computed: {
    sources() {
      return Object.keys(
        this.$store.state.heedy.appSources[this.app.id] || {}
      ).map(id => this.$store.state.heedy.sources[id]);
    }
  },
  watch: {
    app: function(c) {
      this.$store.dispatch("readAppSources", { id: c.id });
    }
  },

  created() {
    this.$store.dispatch("readAppSources", { id: this.app.id });
  }
};
</script>