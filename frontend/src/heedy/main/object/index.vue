<template>
  <component :is="is" :object="object" />
</template>
<script>
import ObjectDefault from "./index_default.vue";

export default {
  props: {
    object: Object,
  },
  head() {
    return {
      title: this.object.name,
    };
  },
  computed: {
    is() {
      if (
        this.$store.state.heedy.object_types[this.object.type] !== undefined
      ) {
        return (
          this.$store.state.heedy.object_types[this.object.type]["index"] ||
          ObjectDefault
        );
      }
      return ObjectDefault;
    },
  },
};
</script>