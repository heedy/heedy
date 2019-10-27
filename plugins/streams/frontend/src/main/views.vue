<template>
  <v-flex style="padding-top: 0px;">
    <v-row>
      <v-col v-for="d in datavis" :key="d.key" cols="12" sm="12" md="6" lg="4">
        <v-card>
          <v-card-title v-if="d.title!==undefined">{{ d.title }}</v-card-title>
          <v-card-text>
            <component :is="view(d.view)" :data="d.data" />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-flex>
</template>
<script>
import ViewNotFound from "./view_notfound.vue";
export default {
  props: {
    source: Object
  },
  data: () => ({
    datavis: null
  }),
  /*
  computed: {
    datavis() {
      let dv = this.$store.state.streams.streams;
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
  },*/
  methods: {
    view(v) {
      let vs = this.$store.state.streams.views;
      if (vs[v] === undefined) {
        return ViewNotFound;
      }
      return vs[v];
    }
  },
  created() {
    this.$app.streams.subscribeQuery(
      this.source,
      "mainviews",
      { i1: -100 },
      dv => {
        let v = Object.keys(dv).map(k => ({ key: k, ...dv[k] }));
        v.sort((a, b) => a.weight - b.weight);
        console.log("datavis", v);
        this.datavis = v;
      }
    );
  },
  beforeDestroy() {
    this.$app.streams.unsubscribeQuery(this.source.id, "mainviews");
  }
};
</script>