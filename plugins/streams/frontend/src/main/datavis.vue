<template>
  <v-flex style="padding-top: 0px;">
    <v-row>
      <v-col v-for="d in datavis" :key="d.key" cols="12" sm="12" md="6" lg="4">
        <v-card>
          <v-card-text>
            <component :is="visualizations[d.component]" />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-flex>
</template>
<script>
export default {
  props: {
    source: Object
  },
  computed: {
    visualizations() {
      return this.$store.state.streams.visualizations;
    },
    datavis() {
      let dv = this.$store.state.streams.datavis;
      if (dv[this.source.id] === undefined) {
        console.log("dataviz null");
        return null;
      }
      dv = dv[this.source.id];
      let v = Object.keys(dv).map(k => ({ key: k, ...dv[k] }));
      v.sort((a, b) => a.weight - b.weight);
      console.log("dataviz", v);
      return v;
    }
  },
  created() {
    this.$app.worker.postMessage("stream_query", { source: this.source });
  }
};
</script>