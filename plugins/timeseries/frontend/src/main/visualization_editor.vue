<template>
  <h-page-container noflex>
    <v-flex
      justify-center
      align-center
      text-center
      style="padding: 10px; padding-bottom: 20px; padding-top: 20px"
    >
      <h1 style="color: #1976d2">Customize Visualization</h1>
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
        <v-card-text>
          <v-text-field label="Name" v-model="name"></v-text-field>
          <h5>Code for Visualization</h5>
          <p>
            Two variables are available: c is the context, which includes the
            query and dataset, and vis is an object that contains the
            configuration for visualizations prepared by heedy and active
            plugins. You can modify visualizations to show by altering the keys
            of vis. You can view detailed documentation here, and go through a
            tutorial here.
          </p>
          <codemirror v-model="code" :options="cmOptions"></codemirror>
        </v-card-text>
        <v-card-actions>
          <v-btn
            v-if="currentVisualizationIndex >= 0"
            dark
            color="red"
            @click="del"
            :loading="loading"
            >Delete</v-btn
          >&nbsp;&nbsp;
          <v-switch v-model="enabled" label="Enabled" />
          <v-spacer></v-spacer>
          <v-btn text @click="close"> Cancel </v-btn>
          <v-btn dark color="blue" @click="save(true)" :loading="loading">Save</v-btn>
        </v-card-actions>
      </v-card>
    </v-flex>
    <v-flex
      justify-center
      align-center
      text-center
      style="padding: 10px; padding-bottom: 20px; padding-top: 30px"
    >
      <h3 style="color: #1976d2">Test Visualization</h3>
    </v-flex>
    <v-flex>
      <v-card>
        <multi-query v-model="query"></multi-query>
        <v-card-actions>
          <v-btn v-if="Object.keys(query).length < 10" text @click="addSeries">
            <v-icon left>add_circle</v-icon>Add Series
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn dark color="blue" @click="runQuery">Test Visualization</v-btn>
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
import MultiQuery from "./dataset/multiquery.vue";
import { deepEqual } from "../../util.mjs";
export default {
  components: {
    MultiQuery,
  },
  data: () => ({
    name: "My Custom Visualization",
    loading: false,
    alert: "",
    enabled: true,
    errmessage: "",
    defaultQuery: {
      "Series 1": {
        timeseries: "",
        t1: "now-3mo",
      },
    },
    query: {
      "Series 1": {
        timeseries: "",
      },
    },
    code: "return vis;",
    cmOptions: {
      tabSize: 2,
      smartIndent: true,
      mode: "text/javascript",
      lineNumbers: true,
      extraKeys: {},
    },
    visquery: {},
  }),

  head: {
    title: "Customize Visualization",
  },
  methods: {
    save: async function (exit) {
      if (this.loading) {
        return;
      }
      this.loading = true;
      try {
        new Function("c", "vis", this.code);
      } catch (e) {
        this.alert = e.message;
        this.loading = false;
        return;
      }

      // Check if the name is unique
      if (
        this.visualizations.find(
          (v, i) => v.name === this.name && i !== this.currentVisualizationIndex
        )
      ) {
        this.alert = "There is already a visualization with this name";
        this.loading = false;
        return;
      }

      const vis = {
        name: this.name,
        enabled: this.enabled,
        code: this.code,
      };

      let visualizations = [...this.visualizations];
      const idx = this.currentVisualizationIndex;

      if (idx === -1) {
        visualizations.push(vis);
      } else {
        visualizations[idx] = vis;
      }

      let result = await this.$frontend.rest(
        "PATCH",
        `api/users/${encodeURIComponent(
          this.$store.state.app.info.user.username
        )}/settings/timeseries`,
        { visualizations: visualizations }
      );

      this.loading = false;

      if (!result.response.ok) {
        this.alert = result.data.error_description;
        return;
      }

      this.alert = "";

      if (exit) {
        this.$router.go(-1);
      } else {
        let qparams = {
          name: this.name,
        };
        if (this.$route.query.q !== undefined) {
          qparams.q = this.$route.query.q;
        }

        if (!deepEqual(this.$route.query, qparams)) {
          this.$router.replace({ query: qparams });
        }
      }
    },
    del: async function () {
      if (this.loading) return;
      this.loading = true;
      if (
        confirm(
          `Are you sure you want to delete the visualization? This cannot be undone. You can disable it instead.`
        )
      ) {
        let visualizations = [...this.visualizations];
        const idx = this.currentVisualizationIndex;
        visualizations.splice(idx, 1);

        let result = await this.$frontend.rest(
          "PATCH",
          `api/users/${encodeURIComponent(
            this.$store.state.app.info.user.username
          )}/settings/timeseries`,
          { visualizations: visualizations }
        );
        this.loading = false;

        if (!result.response.ok) {
          this.alert = result.data.error_description;
          return;
        }

        this.$router.go(-1);
      }
    },
    close() {
      this.$router.go(-1);
    },
    runQuery: async function () {
      let qjson = JSON.stringify(this.query);
      console.vlog("Running query", qjson);
      let qb = btoa(qjson);

      if (this.$route.query.q !== undefined && this.$route.query.q == qb) {
        // The query is identical to current one - manually call process instead of navigating
        this.processQuery(qb);
      } else {
        // Navigate to the query
        this.$router.replace({ query: { ...this.$router.query, q: qb } });
      }
    },
    addSeries() {
      console.vlog("Adding query", this.query);
      for (let i = 1; i < 10; i++) {
        let k = `Series ${i}`;
        if (this.query[k] === undefined) {
          Vue.set(this.query, k, {
            timeseries: "",
            t1: "now-3mo",
          });
          break;
        }
      }
    },
    processQuery(qstring) {
      this.errmessage = "";
      try {
        let qval = atob(qstring);
        let qjson = JSON.parse(qval);
        this.visquery = qjson;
        this.query = JSON.parse(qval); // actually just want a deep copy
      } catch (err) {
        console.error(err);
        this.visquery = {};
        this.query = this.defaultQuery.map((q) => ({ ...q }));
        this.errmessage = "Error reading query";
      }
    },
    getNewName() {
      // First, make sure that the name doesn't overlap any existing visualization
      const name =
        this.$route.query.name !== undefined
          ? this.$route.query.name
          : "My Custom Visualization";
      let curname = name;
      let count = 2;
      while (true) {
        let already_exists = false;
        for (let i = 0; i < this.visualizations.length; i++) {
          if (this.visualizations[i].name === curname) {
            already_exists = true;
            break;
          }
        }
        if (!already_exists) {
          return curname;
        }
        curname = name + " " + count++;
      }
    },
  },
  computed: {
    visualizations() {
      if (
        this.$store.state.app.info.settings.timeseries?.visualizations !==
        undefined
      ) {
        return this.$store.state.app.info.settings.timeseries.visualizations;
      }
      return [];
    },
    currentVisualizationIndex() {
      // If no name is given, or if code IS given, then we are to create a new visualization
      if (
        this.$route.query.name === undefined ||
        this.$route.query.c !== undefined
      ) {
        return -1;
      }

      // Otherwise, find the named visualization
      const name = this.$route.query.name;

      for (let i = 0; i < this.visualizations.length; i++) {
        if (this.visualizations[i].name === name) {
          return i;
        }
      }

      return -1;
    },
    currentVisualization() {
      const i = this.currentVisualizationIndex;
      if (i !== -1) {
        return this.visualizations[i];
      }

      // Otherwise, return what we know about the new visualization from query params
      return {
        name: this.getNewName(),
        code:
          this.$route.query.c !== undefined
            ? this.$route.query.c
            : "return vis;",
        enabled:
          this.$route.query.enabled !== undefined
            ? this.$route.query.enabled
            : true,
      };
    },
  },
  watch: {
    "$route.query": function (n, o) {
      this.errmessage = "";
      if (n.q !== undefined) {
        this.processQuery(n.q);
      } else {
        this.visquery = {};
        this.query = JSON.parse(JSON.stringify(this.defaultQuery));
      }
    },
  },
  created() {
    // If no query, use default
    if (this.$route.query.q !== undefined) {
      this.processQuery(this.$route.query.q);
    }

    this.cmOptions.extraKeys["Ctrl-S"] = () => this.save(false);
    this.cmOptions.extraKeys["Cmd-S"] = () => this.save(false);

    const cv = this.currentVisualization;
    this.code = cv.code;
    this.enabled = cv.enabled;
    this.name = cv.name;
  },
};
</script>
<style>
.CodeMirror {
  border: 1px solid #eee;
  height: auto;
}
</style>