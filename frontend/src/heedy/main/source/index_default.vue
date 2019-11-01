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
    app() {
      if (this.source.app == null) return null;
      if (this.$store.state.heedy.apps == null) return null;
      return this.$store.state.heedy.apps[this.source.app] || null;
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
            if (this.app == null) return false;
            if (this.app.plugin != c.plugin) return false;
          }
          return true;
        }
      );
    }
  },
  watch: {
    source(s) {
      if (s.app != null) {
        let c = this.$store.state.heedy.apps;
        if (c == null || c[s.app] === undefined) {
          this.$store.dispatch("readApp", { id: s.app });
        }
      }
    }
  },
  created() {
    if (this.source.app != null) {
      let c = this.$store.state.heedy.apps;
      if (c == null || c[this.source.app] === undefined) {
        this.$store.dispatch("readApp", { id: this.source.app });
      }
    }
  }
};
</script>