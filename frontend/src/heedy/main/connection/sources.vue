<template>
  <v-flex>
    <h-source-list :sources="sources" />
  </v-flex>
</template>
<script>
export default {
  props: {
    connection: Object
  },
  computed: {
    sources() {
      return Object.keys(
        this.$store.state.heedy.connectionSources[this.connection.id] || {}
      ).map(id => this.$store.state.heedy.sources[id]);
    }
  },
  watch: {
    connection: function(c) {
      this.$store.dispatch("readConnectionSources", { id: c.id });
    }
  },

  created() {
    this.$store.dispatch("readConnectionSources", { id: this.connection.id });
  }
};
</script>