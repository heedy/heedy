<template>
  <h-page-container noflex>
    <component v-for="c in components" :key="c.key" :is="c.component" :source="source" />
  </h-page-container>
</template>
<script>
import { filterComponents } from "../util.js";
export default {
  props: {
    source: Object
  },
  computed: {
    connection() {
      if (this.source.connection == null) return null;
      if (this.$store.state.heedy.connections == null) return null;
      return (
        this.$store.state.heedy.connections[this.source.connection] || null
      );
    },
    components() {
      return filterComponents(
        this.$store.state.heedy.source_components,
        {
          plugin: 4,
          skey: 2,
          type: 1
        },
        c => {
          // Filter out any components that have constraints violated
          if (c.type !== undefined && c.type != this.source.type) return false;
          if (c.skey !== undefined && c.skey != this.source.key) return false;
          if (c.plugin !== undefined) {
            if (this.connection == null) return false;
            if (this.connection.plugin != c.plugin) return false;
          }
          return true;
        }
      );
    }
  },
  watch: {
    source(s) {
      if (s.connection != null) {
        let c = this.$store.state.heedy.connections;
        if (c == null || c[s.connection] === undefined) {
          this.$store.dispatch("readConnection", { id: s.connection });
        }
      }
    }
  },
  created() {
    if (this.source.connection != null) {
      let c = this.$store.state.heedy.connections;
      if (c == null || c[this.source.connection] === undefined) {
        this.$store.dispatch("readConnection", { id: this.source.connection });
      }
    }
  }
};
</script>