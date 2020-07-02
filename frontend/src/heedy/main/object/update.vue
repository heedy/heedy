<template>
  <component :is="is" :object="object" />
</template>
<script>
import ObjectDefault from "./update_default.vue";

export default {
  props: {
    object: Object
  },
  computed: {
    is() {
      if (
        this.$store.state.heedy.object_types[this.object.type] !== undefined
      ) {
        return (
          this.$store.state.heedy.object_types[this.object.type]["update"] ||
          ObjectDefault
        );
      }
      return ObjectDefault;
    }
  }
};
</script>