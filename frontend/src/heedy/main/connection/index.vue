<template>
  <h-page-container noflex>
    <component v-for="c in components" :key="c.key" :is="c.component" :connection="connection" />
  </h-page-container>
</template>
<script>
import { filterComponents } from "../util.js";
export default {
  props: {
    connection: Object
  },
  computed: {
    components() {
      return filterComponents(
        this.$store.state.heedy.connection_components,
        { plugin: 2, type: 1 },
        c => {
          if (c.plugin !== undefined && c.plugin != this.connection.plugin)
            return false;
          if (c.type !== undefined && c.type != this.connection.type)
            return false;
          return true;
        }
      );
    }
  }
};
</script>