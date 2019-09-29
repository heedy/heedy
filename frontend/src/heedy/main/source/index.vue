<template>
  <component :is="is" :source="source" />
</template>
<script>
import SourceDefault from "./index_default.vue";

export default {
  props: {
    source: Object
  },
  computed: {
    is() {
      return (
        this.$store.state.heedy.source_custom_pages[this.source.type] ||
        SourceDefault
      );
    }
  }
};
</script>