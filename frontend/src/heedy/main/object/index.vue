<template>
  <component :is="is" :object="object" />
</template>
<script>
import ObjectDefault from "./index_default.vue";

export default {
  props: {
    object: Object
  },
  computed: {
    is() {
      return (
        this.$store.state.heedy.object_custom_pages[this.object.type] ||
        ObjectDefault
      );
    }
  }
};
</script>