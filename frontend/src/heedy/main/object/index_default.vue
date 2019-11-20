<template>
  <h-page-container noflex>
    <component v-for="c in components" :key="c.key" :is="c.component" :object="object" />
  </h-page-container>
</template>
<script>
import { filterComponents } from "../util.js";
export default {
  props: {
    object: Object
  },
  computed: {
    app() {
      if (this.object.app == null) return null;
      if (this.$store.state.heedy.apps == null) return null;
      return this.$store.state.heedy.apps[this.object.app] || null;
    },
    components() {
      return filterComponents(
        this.$store.state.heedy.object_components,
        {
          plugin: 4,
          skey: 2,
          type: 1
        },
        c => {
          // Filter out any components that have constraints violated
          if (c.type !== undefined && c.type != this.object.type) return false;
          if (c.skey !== undefined && c.skey != this.object.key) return false;
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
    object(s) {
      if (s.app != null) {
        let c = this.$store.state.heedy.apps;
        if (c == null || c[s.app] === undefined) {
          this.$store.dispatch("readApp", { id: s.app });
        }
      }
    }
  },
  created() {
    if (this.object.app != null) {
      let c = this.$store.state.heedy.apps;
      if (c == null || c[this.object.app] === undefined) {
        this.$store.dispatch("readApp", { id: this.object.app });
      }
    }
  }
};
</script>