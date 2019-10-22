<template>
  <h-header
    :icon="source.icon"
    :colorHash="source.id"
    :name="source.name"
    :description="source.description"
  >
    <v-tooltip bottom>
      <template #activator="{on}">
        <v-btn icon v-on="on" :to="`/sources/${source.id}/stream/update`">
          <v-icon>edit</v-icon>
        </v-btn>
      </template>
      <span>Edit Stream</span>
    </v-tooltip>
  </h-header>
</template>
<script>
import api from "../../api.mjs";

export default {
  props: {
    source: Object
  }
};
</script>