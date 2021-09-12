<template>
  <h-header
    :icon="object.icon"
    :defaultIcon="defaultIcon"
    :colorHash="object.id"
    :name="object.name"
    :description="object.description"
    :showTitle="!$slots.default || !$vuetify.breakpoint.xs"
    :toolbar="toolbar"
    :toolbarSize="toolbarSize"
  >
    <slot></slot>
  </h-header>
</template>
<script>
export default {
  props: {
    object: Object,
  },
  computed: {
    defaultIcon() {
      let otype = this.$store.state.heedy.object_types[this.object.type] || {
        icon: "assignment",
      };
      return otype.icon;
    },
    toolbar() {
      // Generate the menu items from the objectMenu
      return Object.values(
        this.$store.state.heedy.objectMenu.reduce(
          (o, m) => ({ ...o, ...m(this.object) }),
          {}
        )
      );
    },
    access() {
      return this.object.access.split(" ");
    },
    toolbarSize() {
      if (this.$vuetify.breakpoint.xs) {
        return 1;
      }
      if (this.$vuetify.breakpoint.sm) {
        return 1;
      }
      if (this.$vuetify.breakpoint.md) {
        return 3;
      }
      return 6;
    },
  },
};
</script>