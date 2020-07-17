<template>
  <h-page-container noflex>
    <v-flex
      justify-center
      align-center
      text-center
      style="padding: 10px; padding-bottom: 20px;padding-top:20px;"
    >
      <h1 style="color:#1976d2;">Data Analysis</h1>
    </v-flex>
    <v-flex>
      <v-card>
        <div style="padding: 10px; padding-bottom: 0;">
          <v-alert
            v-if="alert.length>0"
            text
            outlined
            color="deep-orange"
            icon="error_outline"
          >{{ alert }}</v-alert>
        </div>
        <Query v-model="query"></Query>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn dark color="blue" @click="runQuery" :loading="loading">Run Query</v-btn>
        </v-card-actions>
      </v-card>
    </v-flex>
  </h-page-container>
</template>
<script>
import Query from "./query.vue";
export default {
  components: {
    Query
  },
  data: () => ({
    alert: "",
    query: {
      timeseries: "",
      t1: "now-3mo"
    },
    loading: false
  }),
  methods: {
    runQuery: async function() {
      console.log(JSON.stringify(this.query));
    }
  }
};
</script>