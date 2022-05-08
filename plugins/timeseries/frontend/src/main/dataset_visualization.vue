<template>
  <v-flex style="padding-top: 0px">
    <v-row>
      <slot></slot>
      <v-col
        v-if="datavis.length == 0"
        style="width: 100%; text-align: center"
        cols="12"
        sm="12"
        md="12"
        lg="12"
        xl="12"
      >
        <h1 style="color: #c9c9c9; margin-top: 5%">{{ message }}</h1>
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
          <v-card-title v-if="d.title !== undefined"
            >{{ d.title }}
            <v-spacer />
            <v-menu bottom left>
              <template #activator="{ on, attrs }">
                <v-btn v-on="on" v-bind="attrs" icon>
                  <v-icon>more_vert</v-icon>
                </v-btn>
              </template>
              <v-list>
                <v-list-item @click="showConfig(d)">
                  <v-list-item-icon>
                    <v-icon>code</v-icon>
                  </v-list-item-icon>
                  <v-list-item-content>
                    <v-list-item-title>Configuration</v-list-item-title>
                  </v-list-item-content>
                </v-list-item>
              </v-list>
            </v-menu>
          </v-card-title>
          <v-card-text>
            <component
              :is="visualization(d.visualization)"
              :query="query"
              :config="d.config"
              :data="d.data"
            />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
    <v-dialog v-if="configDialog" v-model="configDialog" max-width="1024px">
      <v-card>
        <v-card-title>
          <span class="headline">{{ configDialogData.title }}</span>
        </v-card-title>
        <v-card-text>
          <codemirror
            :value="JSON.stringify(configDialogData.config, null, '  ')"
            :options="cmOptions"
          ></codemirror>
        </v-card-text>
        <v-card-actions>
          <v-btn outlined @click="customize">Customize</v-btn>
          <div class="flex-grow-1"></div>
          <v-btn color="primary" text @click="configDialog = false"
            >Close</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-flex>
</template>
<script>
import VisNotFound from "./vis_notfound.vue";

const customizationCode = (k,q,c) => `// Only alter the visualization if the queries match
const query = ${JSON.stringify(q, null, "  ")};
// If there is a mismatch, don't modify the visualizations.
if (!c.query.match(query)) return vis;

// If the conditions are met, set the visualization's configuration
vis["${k}"] = ${JSON.stringify(c, null, "  ")};

return vis;
`

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
    configDialog: false,
    configDialogData: {},
    cmOptions: {
      tabSize: 2,
      mode: "text/javascript",
      readOnly: true,
    },
  }),
  methods: {
    visualization(v) {
      let vs = this.$store.state.timeseries.visualizations;
      if (vs[v] === undefined) {
        return VisNotFound;
      }
      return vs[v];
    },
    showConfig(d) {
      this.configDialogData = d;
      this.configDialog = true;
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
    customize() {
      let c = this.configDialogData;
      this.$router.push({path:"/timeseries/customize_visualization", query: {
        name: `Custom ${c.key} visualization`,
        c: customizationCode(c.key, this.query, c.config),
        q: btoa(JSON.stringify(this.query)),
        }});
    }
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
<style>
.CodeMirror {
  border: 1px solid #eee;
  height: auto;
}
</style>