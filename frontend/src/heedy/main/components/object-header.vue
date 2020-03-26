<template>
  <h-header
    :icon="object.icon"
    :defaultIcon="defaultIcon"
    :colorHash="object.id"
    :name="object.name"
    :description="object.description"
    :showTitle="!$vuetify.breakpoint.xs"
  ></h-header>
</template>
<script>
export default {
  props: {
    object: Object
  },
  computed: {
    defaultIcon() {
      let otype = this.$store.state.heedy.object_types[this.object.type] || {
        icon: "assignment"
      };
      return otype.icon;
    }
  }
};
</script>