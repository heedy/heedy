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
            v-if="errmessage.length > 0"
            text
            outlined
            color="deep-orange"
            icon="error_outline"
            >{{ errmessage }}</v-alert
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
        <v-card-text style="margin-top: -20px;margin-bottom: -20px;">
          <h5>Test Query</h5>
        </v-card-text>
        <multi-query v-model="test_query"></multi-query>
        <v-card-actions>
          <v-btn
            v-if="Object.keys(test_query).length < 10"
            text
            @click="addSeries"
          >
            <v-icon left>add_circle</v-icon>Add Series
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn dark color="blue" @click="applyTest">Test Visualization</v-btn>
        </v-card-actions>
      </v-card>
    </v-flex>
    <v-flex v-if="errmessage != ''">
      <div style="width: 100%; text-align: center">
        <h1 style="color: #c9c9c9; margin-top: 5%">{{ viserrmessage }}</h1>
      </div>
    </v-flex>
    <h-dataset-visualization
      v-else
      :query="testVisualization.test_query"
      :user_visualizations="editedVisualizations(testVisualization)"
      :editing_name="testVisualization.name"
    ></h-dataset-visualization>
    <v-flex>
      <v-card>
        <v-card-actions>
          <v-btn
            v-if="index != -1"
            dark
            color="red"
            @click="del"
            :loading="loading"
            >Delete</v-btn
          >&nbsp;&nbsp;
          <v-switch v-model="enabled" label="Enabled" />
          <v-spacer></v-spacer>
          <v-btn text @click="close"> Cancel </v-btn>
          <v-btn
            dark
            color="blue"
            :enabled="isModified"
            @click="save(true)"
            :loading="loading"
            >{{ index===-1? 'Create': 'Save' }}</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-flex>
  </h-page-container>
</template>
<script>
import MultiQuery from "../dataset/multiquery.vue";
import { deepEqual } from "../../../util.mjs";

export default {
  components: {
    MultiQuery,
  },
  props: {
    visualization: Object,
    index: {
      type: Number,
      default: -1,
    }, // The index of visualization that is being edited
  },
  data: () => ({
    loading: false,
    errmessage: "",
    viserrmessage: "",
    code: "",
    name: "",
    enabled: true,
    test_query: {}, // The query that is saved for this visualization
    cmOptions: {
      tabSize: 2,
      smartIndent: true,
      mode: "text/javascript",
      lineNumbers: true,
      extraKeys: {},
    },
  }),
  methods: {
    close() {
      // TODO: Check if the user wants to save the changes
      this.$router.go(-1);
    },
    del: async function () {
      if (this.loading || this.index == -1) return;
      this.loading = true;
      if (
        confirm(
          `Are you sure you want to delete the visualization? This cannot be undone. You might want to disable it instead.`
        )
      ) {
        let visualizations = [...this.visualizations];
        const idx = this.index;
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
          this.errmessage = result.data.error_description;
          return;
        }

        this.$router.go(-1);
      }
    },
    save: async function (exit) {
      if (this.loading) {
        return;
      }
      this.loading = true;
      try {
        new Function("c", "vis", this.code);
      } catch (e) {
        this.errmessage = e.message;
        this.loading = false;
        return;
      }

      // Check if the name is unique
      if (
        this.visualizations.find(
          (v, i) => v.name === this.name && i !== this.index
        )
      ) {
        this.errmessage = "There is already a visualization with this name";
        this.loading = false;
        return;
      }

      // Next check if the current code and query are currently testing
      if (this.code!=this.testVisualization.code || !deepEqual(this.test_query, this.testVisualization.test_query)) {
        if (!confirm("The visualization code or query was changed, but the changes were not tested. Are you sure you want to save without testing?")) {
          this.loading = false;
          return;
        }
      }

      const visualizations = this.editedVisualizations({code: this.code,test_query:this.test_query,enabled: this.enabled,name: this.name});

      const result = await this.$frontend.rest(
        "PATCH",
        `api/users/${encodeURIComponent(
          this.$store.state.app.info.user.username
        )}/settings/timeseries`,
        { visualizations: visualizations }
      );

      this.loading = false;

      if (!result.response.ok) {
        this.errmessage = result.data.error_description;
        return;
      }

      this.errmessage = "";

      if (exit) {
        this.$router.go(-1);
      }
    },
    applyTest: async function () {
      const qjson = JSON.stringify(this.test_query);
      const qb = btoa(qjson);
      const edited = {
        code: this.code,
        test_query: qb
      };
      if (this.name!==this.testVisualization.name) {
        edited.name = this.name;
      }

      this.$router.replace({query: {
        ...this.$router.query,
        ...edited
      }});
    },
    addSeries() {
      for (let i = 1; i < 10; i++) {
        const k = `Series ${i}`;
        if (this.test_query[k] === undefined) {
          Vue.set(this.test_query, k, {
            timeseries: "",
            t1: "now-3mo",
          });
          break;
        }
      }
    },
    parseQuery(qstring) {
      return JSON.parse(atob(qstring));
    },
    ensureUniqueName(name) {
      let curname = name;
      let count = 2;
      while (true) {
        let already_exists = false;
        for (let i = 0; i < this.visualizations.length; i++) {
          if (i !== this.index && this.visualizations[i].name === curname) {
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
    editedVisualizations(vis) {
      const visualizations = [...this.visualizations];

      const idx = this.index;
      if (idx === -1) {
        visualizations.push(vis);
      } else {
        visualizations[idx] = vis;
      }
      return visualizations;
    },
    updateState(visualization, q) {
      console.log(visualization);
      this.errmessage = "";
      this.viserrmessage = "";
      // Initially set all values from the input object
      let code = visualization.code;
      let enabled = visualization.enabled;
      let name = visualization.name;
      let test_query = visualization.test_query;

      // The overwrite them with the query from the route
      if (q.code !== undefined) {
        code = q.code;
      }
      if (q.name !== undefined) {
        name = q.name;
      }
      if (q.enabled !== undefined) {
        enabled = q.enabled === "true";
      }
      if (q.test_query !== undefined) {
        try {
          test_query = this.parseQuery(q.test_query);
        } catch (err) {
          console.error(err);
          this.viserrmessage = "Error reading test query";
        }
      }
      this.code = code;
      this.enabled = enabled;
      this.name = this.ensureUniqueName(name);
      this.test_query =
        test_query !== undefined
          ? test_query
          : {
              "Series 1": {},
            };
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
    testVisualization() {
      // The testState is the state of the visualization as given by the actual vis object
      // and relevant query params
      let name = this.visualization.name;
      let code = this.visualization.code;
      let test_query = this.visualization.test_query;
      test_query =
        test_query !== undefined
          ? test_query
          : {
              "Series 1": {},
            };

      const q = this.$route.query;
      // The overwrite them with the query from the route
      if (q.name!==undefined) {
        name = q.name;
      }
      if (q.code !== undefined) {
        code = q.code;
      }
      if (q.test_query !== undefined) {
        try {
          test_query = this.parseQuery(q.test_query);
        } catch (err) {
          console.error(err);
          this.viserrmessage = "Error reading test query";
        }
      }
      
      return {
        code,test_query,name,enabled: true
      };
    },
    isModified() {
      if (this.index === -1) {
        return true;
      }
      const vis = this.visualization;
      return (
        vis.name !== this.name ||
        vis.enabled !== this.enabled ||
        vis.code !== this.code ||
        deepEqual(vis.test_query, this.test_query) === false
      );
    }
  },
  watch: {
    "$route.query": function (n, o) {
      if (n.code!==undefined && n.code!==o.code) {
        this.code = n.code;
      }
      if (n.enabled!==undefined && n.enabled!==o.enabled) {
        this.enabled = n.enabled === "true";
      }
      if (n.name!==undefined && n.name!==o.name) {
        this.name = n.name;
      }
      if (n.test_query!==undefined && n.test_query!==o.test_query) {
        this.viserrmessage = "";
        try {
          this.test_query = this.parseQuery(n.test_query);
        } catch (err) {
          console.error(err);
          this.viserrmessage = "Error reading test query";
          this.test_query = {};
        }
      }
    },
    visualization(n, o) {
      this.updateState(n, this.$route.query);
    },
  },
  created() {
    this.updateState(this.visualization, this.$route.query);

    this.cmOptions.extraKeys["Ctrl-S"] = () => this.applyTest();
    this.cmOptions.extraKeys["Cmd-S"] = () => this.applyTest();
  },
};
</script>