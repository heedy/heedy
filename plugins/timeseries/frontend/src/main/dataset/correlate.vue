<template>
  <v-layout row>
    <v-flex sm6 xs12 :md7="!showTransform" :md4="showTransform">
      <h-object-picker v-model="timeseries" type="timeseries" multiple></h-object-picker>
    </v-flex>
    <v-flex v-if="showTransform && !$vuetify.breakpoint.sm" md4 xs12>
      <transform v-model="transform"></transform>
    </v-flex>
    <v-flex sm6 xs12 :md5="!showTransform" :md4="showTransform">
      <div style="display: grid;grid-template-columns: auto min-content;grid-gap:0px">
        <interpolator v-model="interpolator"></interpolator>
        <div style="margin:auto; padding-left:10px;">
          <v-menu offset-y>
            <template v-slot:activator="{ on: menu, attrs }">
              <v-tooltip bottom>
                <template v-slot:activator="{ on: tooltip }">
                  <v-btn outlined text v-bind="attrs" v-on="{ ...tooltip, ...menu }">{{ colname }}</v-btn>
                </template>
                <span>Column Options</span>
              </v-tooltip>
            </template>
            <v-list>
              <v-list-item @click="toggleTransform">
                <v-list-item-icon>
                  <v-icon>code</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title v-if="!showTransform">Add Transform</v-list-item-title>
                  <v-list-item-title v-else>Remove Transform</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="showDialog">
                <v-list-item-icon>
                  <v-icon>create</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>Rename Column</v-list-item-title>
                </v-list-item-content>
              </v-list-item>
              <v-list-item @click="()=> $emit('delete')">
                <v-list-item-icon>
                  <v-icon>delete</v-icon>
                </v-list-item-icon>
                <v-list-item-content>
                  <v-list-item-title>Remove</v-list-item-title>
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
    <v-dialog v-model="dialog" max-width="500">
      <v-card>
        <v-card-title class="headline grey lighten-2" primary-title>Rename Column</v-card-title>

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
  </v-layout>
</template>
<script>
import Interpolator from "./interpolator.vue";
import Transform from "./transform.vue";
export default {
  components: {
    Interpolator,
    Transform
  },
  props: {
    colname: String,
    value: Object
  },
  data: () => ({
    dialog: false,
    coltext: "",
    wantTransform: false
  }),
  methods: {
    showDialog() {
      this.coltext = this.colname;
      this.dialog = true;
    },
    setCol() {
      this.dialog = false;
      this.$emit("update:colname", this.coltext);
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
    }
  },
  computed: {
    showTransform() {
      return this.transform != "" || this.wantTransform;
    },
    transform: {
      get() {
        return this.value.transform || "";
      },
      set(t) {
        this.$emit("input", { ...this.value, transform: t });
      }
    },
    interpolator: {
      get() {
        return this.value.interpolator || "closest";
      },
      set(v) {
        this.$emit("input", { ...this.value, interpolator: v });
      }
    },
    timeseries: {
      get() {
        if (this.value.merge !== undefined) {
          return this.value.merge.map(v => v.timeseries);
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
          merge: v.map(e => ({
            timeseries: e
          }))
        });
      }
    }
  }
};
</script>