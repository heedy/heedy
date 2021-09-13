<template>
  <h-header
    :icon="app.icon"
    defaultIcon="settings_input_component"
    :colorHash="app.id"
    :name="app.name"
    :description="app.enabled ? app.description : '(disabled)'"
    :toolbar="toolbar"
    :toolbarSize="toolbarSize"
  >
  </h-header>
</template>

<script>
export default {
  props: {
    app: Object,
  },
  watch: {
    showkey(newv) {
      this.token = "...";
    },
  },
  computed: {
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
    toolbar() {
      return Object.values(
        this.$store.state.heedy.appMenu.reduce(
          (o, m) => ({ ...o, ...m(this.app) }),
          {}
        )
      );
    },
  },
};
</script>