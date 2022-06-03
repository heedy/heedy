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
                <v-list-item @click="customize(d)">
                  <v-list-item-icon>
                    <v-icon>code</v-icon>
                  </v-list-item-icon>
                  <v-list-item-content>
                    <v-list-item-title>Customize</v-list-item-title>
                  </v-list-item-content>
                </v-list-item>
              </v-list>
            </v-menu>
          </v-card-title>
          <v-card-text>
            <component
              :is="getVisComponentByType(d.type)"
              :query="query"
              :config="d.config"
              :data="d.data"
              :type="d.type"
            />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-flex>
</template>
<script>
import VisNotFound from "./vis_notfound.vue";

const customizationCode = (k,q,c) => `// The c variable holds context data, and vis holds visualizations
// Only customize the visualization for queries that match this one
const q = ${JSON.stringify(q, null, "  ")};
if (!c.query.isEqual(q)) return vis;

// If the conditions are met, set the visualization's configuration
vis[${JSON.stringify(k)}] = ${JSON.stringify(c, null, "  ")};

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

function ValidQuery(q) {
  for (let k of Object.keys(q)) {
    const e = q[k];
    if (e.dt===undefined && (e.timeseries===undefined || e.timeseries.length==0)) return false;
  }
  return true;
}

export default {
  props: {
    query: Object,
    live: {
      type: Boolean,
      default: true,
    },
    user_visualizations: {
      type: Array,
      default: null
    },
  },
  data: () => ({
    message: "Querying Data...",
    datavis: [],
    qkey: "",
    cmOptions: {
      tabSize: 2,
      mode: "text/javascript",
      readOnly: true,
    },
  }),
  methods: {
    getVisComponentByType(v) {
      let vs = this.$store.state.timeseries.visualizationTypes;
      if (vs[v] === undefined) {
        return VisNotFound;
      }
      return vs[v];
    },
    subscribe(q,uv) {
      if (this.qkey != "") {
        this.$frontend.timeseries.unsubscribeQuery(this.qkey);
        this.qkey = "";
      }
      q = CleanQuery(q);
      if (!ValidQuery(q)) {
        this.message = "Empty Query";
        return;
      }
      this.message = "Loading...";
      this.datavis = [];
      this.qkey = this.$frontend.timeseries.subscribeQuery(
        q,
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
            v.map((vi) => `${vi.key} (${vi.type})`)
          );
          this.datavis = v;
          this.message = "No Data";
        },
        uv
      );
    },
    customize(c) {
      // Create a special title. If the query is a single timeseries,
      // name the visualization after the timeseries (if available)
      let name = `Custom ${c.key} visualization`;
      const keys = Object.keys(this.query);
      if (keys.length==1) {
        const k = keys[0];
        const q = this.query[k];
        if (q.dt === undefined && q.timeseries !== undefined && typeof q.timeseries === "string") {
          if (this.$store.state.heedy.objects[q.timeseries]!==undefined) {
            name = `Custom ${this.$store.state.heedy.objects[q.timeseries].name} ${c.key} visualization`;
          }
        }
      }
      this.$router.push({path:"/timeseries/visualization/create", query: {
        name: name,
        code: customizationCode(c.key,CleanQuery(this.query),{
          title: c.title,
          weight: c.weight,
          type: c.type,
          config: c.config,
        }),
        test_query: btoa(JSON.stringify(this.query)),
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
        this.subscribe(n,this.user_visualizations);
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
      this.subscribe(this.query,this.user_visualizations);
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