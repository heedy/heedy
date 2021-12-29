<template>
  <v-container fluid>
    <v-layout row>
      <v-flex sm6 xs12 :md7="!showTransform" :md4="showTransform">
        <h-duration-editor label="Each" v-if="tsquery" v-model="dt" />
        <h-object-picker
          v-else
          v-model="timeseries"
          type="timeseries"
          multiple
        ></h-object-picker>
      </v-flex>
      <v-flex v-if="showTransform && !$vuetify.breakpoint.sm" md4 xs12>
        <transform v-model="transform"></transform>
      </v-flex>
      <v-flex sm6 xs12 :md5="!showTransform" :md4="showTransform">
        <div
          style="
            display: grid;
            grid-template-columns: auto min-content;
            grid-gap: 0px;
          "
        >
          <div style="overflow: hidden">
            <h-timeseries-range-picker
              :value="value"
              @input="(v) => $emit('input', v)"
            />
          </div>
          <div style="margin: auto; padding-left: 10px">
            <v-menu offset-y>
              <template v-slot:activator="{ on: menu, attrs }">
                <v-tooltip bottom>
                  <template v-slot:activator="{ on: tooltip }">
                    <v-btn
                      :icon="!isDataset || tsquery"
                      :outlined="isDataset && !tsquery"
                      :text="isDataset && !tsquery"
                      v-bind="attrs"
                      v-on="{ ...tooltip, ...menu }"
                    >
                      <v-icon v-if="!isDataset || tsquery">more_vert</v-icon>
                      <template v-else>{{ colname }}</template>
                    </v-btn>
                  </template>
                  <span>Options</span>
                </v-tooltip>
              </template>
              <v-list>
                <v-list-item @click="toggleTransform" v-if="!tsquery">
                  <v-list-item-icon>
                    <v-icon>code</v-icon>
                  </v-list-item-icon>
                  <v-list-item-content>
                    <v-list-item-title v-if="!showTransform"
                      >Add Transform</v-list-item-title
                    >
                    <v-list-item-title v-else
                      >Remove Transform</v-list-item-title
                    >
                  </v-list-item-content>
                </v-list-item>
                <v-list-item v-if="isDataset && !tsquery" @click="showDialog">
                  <v-list-item-icon>
                    <v-icon>create</v-icon>
                  </v-list-item-icon>
                  <v-list-item-content>
                    <v-list-item-title>Rename Column</v-list-item-title>
                  </v-list-item-content>
                </v-list-item>
                <v-list-item @click="toggleTSQ">
                  <v-list-item-icon>
                    <v-icon>query_builder</v-icon>
                  </v-list-item-icon>
                  <v-list-item-content>
                    <v-list-item-title v-if="tsquery"
                      >Switch to Timeseries</v-list-item-title
                    >
                    <v-list-item-title v-else
                      >Switch to Time Range</v-list-item-title
                    >
                  </v-list-item-content>
                </v-list-item>
                <v-list-item
                  v-if="removeSeries"
                  @click="() => $emit('remove-series')"
                >
                  <v-list-item-icon>
                    <v-icon>remove_circle</v-icon>
                  </v-list-item-icon>
                  <v-list-item-content>
                    <v-list-item-title>Remove Series</v-list-item-title>
                  </v-list-item-content>
                </v-list-item>
              </v-list>
            </v-menu>
          </div>
        </div>
      </v-flex>
      <v-flex v-if="showTransform && $vuetify.breakpoint.sm" sm12>
        <transform v-model="transform"></transform>
      </v-flex>
      <v-flex v-for="(v, k) in dataset" :key="k" lg12 md12 sm12 xs12 xl12>
        <div style="display: block">
          <correlate
            :colname="k"
            @update:colname="(cn) => updateCol(k, cn)"
            :value="v"
            @input="(val) => setDatasetVal(k, val)"
            @delete="() => delCol(k)"
          ></correlate>
        </div>
      </v-flex>
      <v-flex
        lg12
        md12
        sm12
        xs12
        style="
          padding-top: 0;
          padding-bottom: 0;
          margin-top: -20px;
          margin-bottom: 0;
          padding-left: 8px;
        "
      >
        <v-btn text @click="addDatasetElement">
          <v-icon left>add</v-icon>Correlate With...
        </v-btn>
      </v-flex>
    </v-layout>
    <v-dialog v-model="dialog" max-width="500">
      <v-card>
        <v-card-title class="headline grey lighten-2" primary-title
          >Rename Column</v-card-title
        >

        <v-card-text>
          <v-text-field autofocus v-model="coltext" />
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="secondary" text @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" text @click="setCol">Set</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>
<script>
import Transform from "./transform.vue";
import Correlate from "./correlate.vue";
export default {
  props: {
    value: Object,
    removeSeries: {
      type: Boolean,
      default: false,
    },
  },
  components: {
    Transform,
    Correlate,
  },
  data: () => ({
    dialog: false,
    wantTransform: false,
    coltext: "",
  }),
  methods: {
    showDialog() {
      this.coltext = this.colname;
      this.dialog = true;
    },
    setCol() {
      this.dialog = false;
      if (this.hasCol(this.coltext)) {
        console.error("Column already exists");
        return;
      }
      this.$emit("input", {
        ...this.value,
        key: this.coltext,
      });
    },
    setDatasetVal(k, v) {
      let datasetv = {
        ...this.value.dataset,
      };
      datasetv[k] = v;

      this.$emit("input", { ...this.value, dataset: datasetv });
    },
    updateCol(cur, next) {
      if (this.hasCol(next)) {
        console.error("Column already exists");
        return;
      }
      let datasetv = {
        ...this.value.dataset,
      };
      datasetv[next] = datasetv[cur];
      if (next != cur) {
        delete datasetv[cur];
      }
      this.$emit("input", { ...this.value, dataset: datasetv });
    },
    hasCol(v) {
      return this.colname == v || this.dataset[v] !== undefined;
    },
    delCol(col) {
      let datasetv = {
        ...this.value.dataset,
      };
      delete datasetv[col];
      this.$emit("input", { ...this.value, dataset: datasetv });
    },
    addDatasetElement() {
      let newElement = { timeseries: "" };
      // Check if it can be one of x,y,z
      let startWith = ["y", "z"];
      for (let i = 0; i < startWith.length; i++) {
        if (!this.hasCol(startWith[i])) {
          let ds = {
            ...this.dataset,
          };
          ds[startWith[i]] = newElement;
          this.$emit("input", { ...this.value, dataset: ds });
          return;
        }
      }
      // If not, start from beginning of alphabet
      for (let i = 0; i < 26; i++) {
        let curval = String.fromCharCode("a".charCodeAt(0) + i);
        if (!this.hasCol(curval)) {
          let ds = {
            ...this.dataset,
          };
          ds[curval] = newElement;
          this.$emit("input", { ...this.value, dataset: ds });
          return;
        }
      }

      // Annnd we ran out of letters.
      // TODO: don't fail silently?
      console.error("Ran out of letters for correlates!");
    },
    toggleTransform() {
      if (this.wantTransform || this.transform != "") {
        this.wantTransform = false;
        let vv = { ...this.value };
        delete vv.transform;
        this.$emit("input", vv);
        return;
      }
      this.wantTransform = true;
    },
    toggleTSQ() {
      let nv = { ...this.value };
      if (this.tsquery) {
        delete nv.dt;
        nv.timeseries = "";
      } else {
        delete nv.merge;
        delete nv.timeseries;
        delete nv.transform;
        delete nv.key;
        this.wantTransform = false;
        nv.dt = 60 * 60;
        if (nv.dataset === undefined || Object.keys(nv.dataset).length == 0) {
          nv.dataset = {
            y: { timeseries: "" },
          };
        }
      }

      this.$emit("input", nv);
    },
  },
  computed: {
    dt: {
      get() {
        return this.value.dt;
      },
      set(v) {
        this.$emit("input", { ...this.value, dt: v });
      },
    },
    tsquery() {
      return this.value.dt !== undefined;
    },
    showTransform() {
      return this.transform != "" || this.wantTransform;
    },
    transform: {
      get() {
        return this.value.transform || "";
      },
      set(t) {
        this.$emit("input", { ...this.value, transform: t });
      },
    },
    colname() {
      return this.value.key || "x";
    },
    isDataset() {
      return (
        this.value.dataset !== undefined &&
        Object.keys(this.value.dataset).length > 0
      );
    },
    timeseries: {
      get() {
        if (this.value.merge !== undefined) {
          return this.value.merge.map((v) => v.timeseries);
        }
        if (
          this.value.timeseries !== undefined &&
          this.value.timeseries != ""
        ) {
          return [this.value.timeseries];
        }
        return [];
      },
      set(v) {
        let curval = { ...this.value };
        delete curval.timeseries;
        delete curval.merge;

        if (typeof v === "string") {
          v = [v];
        }
        if (v.length == 1) {
          this.$emit("input", { ...curval, timeseries: v[0] });
          return;
        }
        this.$emit("input", {
          ...curval,
          merge: v.map((e) => ({
            timeseries: e,
          })),
        });
      },
    },
    dataset() {
      return this.value.dataset || {};
    },
  },
};
</script>