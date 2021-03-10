<template>
  <v-flex style="padding-top: 0px">
    <v-row>
      <slot>
        <v-col
          v-if="datavis.length == 0"
          style="width: 100%; text-align: center"
        >
          <h1 style="color: #c9c9c9; margin-top: 5%">{{ message }}</h1>
        </v-col>
      </slot>
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
            <component
              :is="visualization(d.visualization)"
              :query="query"
              :config="d.config"
            />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-flex>
</template>
<script>
import VisNotFound from "./vis_notfound.vue";

function CleanQuery(q) {
  let q2 = {};
  Object.keys(q).forEach((k) => {
    let e = q[k];
    let e2 = {
      ...q[k],
    };

    if (e.i1 !== undefined && !isNaN(e.i1)) {
      e2.i1 = parseInt(e.i1);
    }
    if (e.i2 !== undefined && !isNaN(e.i2)) {
      e2.i2 = parseInt(e.i2);
    }
    if (e.limit !== undefined && !isNaN(e.limit)) {
      e2.limit = parseInt(e.limit);
    }
    if (e.i !== undefined && !isNaN(e.i)) {
      e2.i = parseInt(e.i);
    }
    q2[k] = e2;
  });
  return q2;
}

export default {
  props: {
    query: Object,
    live: {
      type: Boolean,
      default: true,
    },
  },
  data: () => ({
    message: "Querying Data...",
    datavis: [],
    qkey: "",
  }),
  methods: {
    visualization(v) {
      let vs = this.$store.state.timeseries.visualizations;
      if (vs[v] === undefined) {
        return VisNotFound;
      }
      return vs[v];
    },
    subscribe(q) {
      if (this.qkey != "") {
        this.$frontend.timeseries.unsubscribeQuery(this.qkey);
        this.qkey = "";
      }
      this.message = "Loading...";
      this.datavis = [];
      this.qkey = this.$frontend.timeseries.subscribeQuery(
        CleanQuery(q),
        (dv) => {
          if (dv.status !== undefined) {
            // Special-case query status messages
            this.message = dv.status;
            return;
          }

          dv = dv.visualizations;

          let v = Object.keys(dv).map((k) => ({ key: k, ...dv[k] }));
          v.sort((a, b) => a.weight - b.weight);
          console.vlog(
            "Received visualizations:",
            v.map((vi) => `${vi.key} (${vi.visualization})`)
          );
          this.datavis = v;
          this.message = "No Data";
        }
      );
    },
  },
  watch: {
    query(n, o) {
      if (this.qkey != "") {
        this.$frontend.timeseries.unsubscribeQuery(this.qkey);
        this.qkey = "";
      }
      if (Object.keys(n).length > 0) {
        this.subscribe(n);
      } else {
        this.datavis = [];
        this.message = "";
      }
    },
    live(n, o) {},
  },
  created() {
    // Only subscribe if non-empty query, or modify the query to be the default
    if (Object.keys(this.query).length > 0) {
      this.subscribe(this.query);
    } else {
      this.message = "";
    }
  },
  beforeDestroy() {
    if (this.qkey != "") {
      this.$frontend.timeseries.unsubscribeQuery(this.qkey);
    }
  },
};
</script>
