<template>
  <h-header
    :icon="object.icon"
    :defaultIcon="defaultIcon"
    :colorHash="object.id"
    :name="object.name"
    :description="object.description"
    :showTitle="!$vuetify.breakpoint.xs"
  >
    <slot></slot>
    <v-tooltip bottom v-if="access.includes('*') || access.includes('write')">
      <template #activator="{ on }">
        <v-btn icon v-on="on" :to="`/objects/${object.id}/update`">
          <v-icon>edit</v-icon>
        </v-btn>
      </template>
      <span>Edit</span>
    </v-tooltip>
    <v-tooltip v-if="object.app != null && !$vuetify.breakpoint.xs" bottom>
      <template #activator="{ on }">
        <v-btn icon :to="`/apps/${object.app}`" v-on="on">
          <h-icon
            :image="app.icon"
            :defaultIcon="defaultIcon"
            :colorHash="app.id"
            :size="30"
          ></h-icon
        ></v-btn>
      </template>
      <span>{{ app.name }}</span>
    </v-tooltip>
  </h-header>
</template>
<script>
export default {
  props: {
    object: Object,
  },
  watch: {
    object(newobj, oobj) {
      if (newobj.app != null && newobj.app != oobj.app) {
        this.$store.dispatch("readApp", {
          id: newobj.app,
        });
      }
    },
  },
  computed: {
    defaultIcon() {
      let otype = this.$store.state.heedy.object_types[this.object.type] || {
        icon: "assignment",
      };
      return otype.icon;
    },
    access() {
      return this.object.access.split(" ");
    },
    app() {
      let empty_app = {
        id: this.object.id,
        icon: "settings_input_component",
        name: "Go to app",
      };
      if (this.$store.state.heedy.apps == null) {
        return empty_app;
      }
      return this.$store.state.heedy.apps[this.object.app] || empty_app;
    },
  },
  created() {
    if (this.object.app != null) {
      this.$store.dispatch("readApp", {
        id: this.object.app,
      });
    }
  },
};
</script>