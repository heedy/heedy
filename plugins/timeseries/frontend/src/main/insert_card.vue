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
      <h-timeseries-datapoint-inserter
        :object="object"
      ></h-timeseries-datapoint-inserter>
    </v-card-text>
  </v-card>
</template>
<script>
export default {
  props: {
    object: Object,
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