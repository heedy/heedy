<template>
  <v-flex style="padding-top: 0px;">
    <v-row>
      <v-col
        v-if="datavis.length == 0"
        style="width: 100%; text-align: center;"
      >
        <h1 style="color: #c9c9c9;margin-top: 5%;">{{ message }}</h1>
      </v-col>
      <v-col
        v-for="d in datavis"
        :key="d.key"
        cols="12"
        sm="12"
        md="6"
        lg="6"
        xl="4"
      >
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
    object: Object,
  },
  data: () => ({
    message: "Querying Data...",
    datavis: null,
    subscribed: false,
    defaultQuery: { i1: -1000 },
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
        this.$frontend.timeseries.unsubscribeQuery(this.object.id, "mainviews");
      }
      this.subscribed = true;
      this.message = "Querying Data...";
      this.datavis = [];
      this.$frontend.timeseries.subscribeQuery(
        this.object,
        "mainviews",
        q,
        (dv) => {
          if (dv.query_status !== undefined) {
            // Special-case query status messages
            this.message = dv.query_status.data;
            return;
          }
          let v = Object.keys(dv).map((k) => ({ key: k, ...dv[k] }));
          v.sort((a, b) => a.weight - b.weight);
          console.log(
            "Received views:",
            v.map((vi) => `${vi.key} (${vi.view})`)
          );
          this.datavis = v;
          this.message = "No Data";
        }
      );
    },
  },
  watch: {
    "$route.query": function(n, o) {
      this.subscribe(n);
    },
    object(n, o) {
      if (n.id != o.id) {
        if (this.subscribed) {
          this.$frontend.timeseries.unsubscribeQuery(o.id, "mainviews");
          this.subscribed = false;
          this.subscribe(this.$route.query);
        }
      }
    },
  },
  created() {
    // Only subscribe if non-empty query, or modify the query to be the default
    if (Object.keys(this.$route.query).length > 0) {
      this.subscribe(this.$route.query);
    } else {
      this.$router.replace({ query: this.defaultQuery });
    }
  },
  beforeDestroy() {
    this.$frontend.timeseries.unsubscribeQuery(this.object.id, "mainviews");
  },
};
</script>
