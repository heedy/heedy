<template>
  <v-flex style="padding-top: 0px;">
    <v-row>
      <v-col v-for="d in datavis" :key="d.key" cols="12" sm="12" md="6" lg="4">
        <v-card>
          <v-card-title v-if="d.title !== undefined">{{
            d.title
          }}</v-card-title>
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
    object: Object
  },
  data: () => ({
    datavis: null,
    subscribed: false
  }),
  /*
  computed: {
    datavis() {
      let dv = this.$store.state.timeseries.timeseries;
      if (dv[this.object.id] === undefined) {
        console.log("dataviz null");
        return null;
      }
      dv = dv[this.object.id];
      let v = Object.keys(dv).map(k => ({ key: k, ...dv[k] }));
      v.sort((a, b) => a.weight - b.weight);
      console.log("dataviz", v);
      return v;
    }
  },*/
  methods: {
    view(v) {
      let vs = this.$store.state.timeseries.views;
      if (vs[v] === undefined) {
        return ViewNotFound;
      }
      return vs[v];
    },
    subscribe(q) {
      if (this.subscribed) {
        this.$app.timeseries.unsubscribeQuery(this.object.id, "mainviews");
      }
      this.subscribed = true;
      this.$app.timeseries.subscribeQuery(this.object, "mainviews", q, dv => {
        let v = Object.keys(dv).map(k => ({ key: k, ...dv[k] }));
        v.sort((a, b) => a.weight - b.weight);
        console.log("datavis", v);
        this.datavis = v;
      });
    }
  },
  watch: {
    "$route.query": function(n, o) {
      if (this.subscribed) {
        this.$app.timeseries.unsubscribeQuery(this.object.id, "mainviews");
      }
      this.subscribe(n);
    },
    object(n, o) {
      if (n.id != o.id) {
        if (this.subscribed) {
          this.$app.timeseries.unsubscribeQuery(this.object.id, "mainviews");
          this.subscribed = false;
          this.subscribe(this.$route.query);
        }
      }
    }
  },
  created() {
    // Only subscribe if non-empty query, since the header will fire a default query if it is empty
    if (Object.keys(this.$route.query).length > 0) {
      this.subscribe(this.$route.query);
    }
  },
  beforeDestroy() {
    this.$app.timeseries.unsubscribeQuery(this.object.id, "mainviews");
  }
};
</script>
