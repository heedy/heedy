<template>
  <v-card>
    <v-card-title>
      <router-link
        :to="`/objects/${object.id}`"
        style="text-decoration: inherit; color: inherit"
      >
        <h-icon
          :image="object.icon"
          :defaultIcon="defaultIcon"
          :colorHash="object.id"
          :size="30"
        ></h-icon>
        <span style="padding-left: 10px">{{ object.name }}</span>
      </router-link>
      <v-spacer />
      <h-app-button :appid="object.app" :size="20" />
    </v-card-title>
    <v-card-text>
      <inserter :object="object"></inserter>
    </v-card-text>
  </v-card>
</template>
<script>
import Inserter from "./insert.vue";
export default {
  props: {
    object: Object,
  },
  components: {
    Inserter,
  },
  computed: {
    defaultIcon() {
      return (
        this.$store.state.heedy.object_types["timeseries"].icon ||
        "brightness_1"
      );
    },
  },
};
</script>