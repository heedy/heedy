<template>
  <h-page-container noflex>
    <component
      v-for="c in components"
      :key="c.key"
      :is="c.component"
      :app="app"
    />
  </h-page-container>
</template>
<script>
import { filterComponents } from "../util.js";
export default {
  props: {
    app: Object,
  },
  head() {
    return {
      title: this.app.name,
    };
  },
  computed: {
    components() {
      return filterComponents(
        this.$store.state.heedy.app_components,
        { plugin: 2, type: 1 },
        (c) => {
          if (c.plugin !== undefined && c.plugin != this.app.plugin)
            return false;
          if (c.type !== undefined && c.type != this.app.type) return false;
          return true;
        }
      );
    },
  },
};
</script>