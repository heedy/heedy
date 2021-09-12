<template>
  <v-tooltip v-if="!isList" bottom>
    <template #activator="{ on }">
      <v-btn icon v-on="on" @click="runAnalysis">
        <v-icon>fas fa-chart-bar</v-icon>
      </v-btn>
    </template>
    <span>Analysis</span>
  </v-tooltip>
  <v-list-item v-else @click="runAnalysis">
    <v-list-item-icon>
      <v-icon>fas fa-chart-bar</v-icon>
    </v-list-item-icon>
    <v-list-item-content>
      <v-list-item-title>Analysis</v-list-item-title>
    </v-list-item-content>
  </v-list-item>
</template>
<script>
export default {
  props: {
    objectid: String,
    isList: Boolean,
  },
  methods: {
    runAnalysis() {
      let qjson = JSON.stringify({
        "Series 1": {
          timeseries: this.objectid,
          ...this.$route.query,
        },
      });
      let qb = btoa(qjson);
      this.$router.push({ path: "/timeseries/dataset", query: { q: qb } });
    },
  },
};
</script>
