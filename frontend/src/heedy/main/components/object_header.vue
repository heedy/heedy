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
    <h-app-button
      v-if="!$vuetify.breakpoint.xs"
      :appid="object.app"
      :size="30"
    />
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
    access() {
      return this.object.access.split(" ");
    },
  },
};
</script>