<template>
  <h-page-container noflex>
    <v-flex
      justify-center
      align-center
      text-center
      style="padding: 10px; padding-bottom: 20px; padding-top: 20px"
    >
      <h1 style="color: #1976d2">Data Analysis</h1>
    </v-flex>
    <v-flex>
      <v-card>
        <div style="padding: 10px; padding-bottom: 0">
          <v-alert
            v-if="alert.length > 0"
            text
            outlined
            color="deep-orange"
            icon="error_outline"
            >{{ alert }}</v-alert
          >
        </div>
        <multi-query v-model="query"></multi-query>
        <v-card-actions>
          <v-btn text @click="addSeries">
            <v-icon left>add_circle</v-icon>Add Series
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn dark color="blue" @click="runQuery" :loading="loading"
            >Run Query</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-flex>
    <v-flex v-if="errmessage != ''">
      <div style="width: 100%; text-align: center">
        <h1 style="color: #c9c9c9; margin-top: 5%">{{ errmessage }}</h1>
      </div>
    </v-flex>
    <h-dataset-visualization v-else :query="visquery"></h-dataset-visualization>
  </h-page-container>
</template>
<script>
import MultiQuery from "./multiquery.vue";
export default {
  components: {
    MultiQuery,
  },
  data: () => ({
    alert: "",
    defaultQuery: [
      {
        timeseries: "",
        t1: "now-3mo",
      },
    ],
    query: [
      {
        timeseries: "",
        t1: "now-3mo",
      },
    ],
    visquery: [],
    loading: false,
    errmessage: "",
  }),
  methods: {
    runQuery: async function () {
      let qjson = JSON.stringify(this.query);
      console.vlog("Running query", qjson);
      let qb = btoa(qjson);

      if (this.$route.query.q !== undefined && this.$route.query.q == qb) {
        // The query is identical to current one - manually call process instead of navigating
        this.processQuery(qb);
      } else {
        // Navigate to the query
        this.$router.replace({ query: { q: qb } });
      }
    },
    addSeries() {
      this.query.push({
        timeseries: "",
        t1: "now-3mo",
      });
    },
    processQuery(qstring) {
      this.errmessage = "";
      try {
        let qval = atob(qstring);
        let qjson = JSON.parse(qval);
        this.visquery = qjson;
        this.query = qjson.map((q) => ({ ...q }));
      } catch (err) {
        console.error(err);
        this.visquery = [];
        this.query = this.defaultQuery.map((q) => ({ ...q }));
        this.errmessage = "Error reading query";
      }
    },
  },
  watch: {
    "$route.query": function (n, o) {
      this.errmessage = "";
      if (n.q !== undefined) {
        this.processQuery(n.q);
      } else {
        this.visquery = [];
        this.query = this.defaultQuery.map((q) => ({ ...q }));
      }
    },
  },
  created() {
    // If no query, use default
    if (this.$route.query.q !== undefined) {
      this.processQuery(this.$route.query.q);
    }
  },
};
</script>
