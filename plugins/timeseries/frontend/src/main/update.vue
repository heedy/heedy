<template>
  <h-object-updater :object="object" :meta="meta" :validator="validate">
    <v-container style="margin-top: -30px; margin-bottom: -20px">
      <v-row>
        <v-flex style="margin: auto; flex: 0 0 8em">
          <h3>Data Type:</h3>
        </v-flex>
        <v-flex style="flex: 1 1">
          <v-select
            v-model="curtype"
            :items="[...types, { title: 'Custom', key: 'custom' }]"
            item-text="title"
            item-value="key"
          />
        </v-flex>
      </v-row>
      <v-row v-if="curtype == 'custom'">
        <v-flex>
          <h-schema-editor v-model="schema" />
        </v-flex>
      </v-row>
      <v-row
        v-else-if="$store.state.timeseries.types[curtype].editor !== undefined"
      >
        <v-flex>
          <component
            :is="$store.state.timeseries.types[curtype].editor"
            v-model="schema"
          />
        </v-flex>
      </v-row>
    </v-container>
  </h-object-updater>
</template>
<script>
import { deepEqual } from "../../util.mjs";
import Validator from "../../dist/json-schema.mjs";
export default {
  props: {
    object: Object,
  },
  data: () => ({
    meta: {},
    ct: "custom",
  }),
  methods: {
    validate(o) {
      if (o.meta !== undefined && o.meta.schema !== undefined) {
        if (o.meta.schema == null) {
          return "Invalid Schema";
        }
      }
      return "";
    },
  },
  computed: {
    types() {
      return Object.values(this.$store.state.timeseries.types);
    },
    curtype: {
      get() {
        return this.ct;
      },
      set(ct) {
        if (ct != "custom") {
          this.schema = this.$store.state.timeseries.types[ct].schema;
        }
        this.ct = ct;
      },
    },
    schema: {
      get() {
        if (this.meta.schema !== undefined) {
          return this.meta.schema;
        }
        return this.object.meta.schema;
      },
      set(s) {
        if (deepEqual(s, this.object.meta.schema)) {
          // Remove the modification
          let m = {};
          Object.keys(this.meta).forEach((k) => {
            if (k != "schema") {
              m[k] = this.meta[k];
            }
          });
          this.meta = m;
        }
        this.meta = { ...this.meta, schema: s };
      },
    },
  },
  created() {
    // Set up the correct type
    for (let i = 0; i < this.types.length; i++) {
      let t = this.types[i];
      if (t.meta !== undefined) {
        if (new Validator(t.meta).validate(this.object.meta.schema).valid) {
          this.ct = t.key;
          break;
        }
      } else if (deepEqual(this.object.meta.schema, t.schema)) {
        this.ct = t.key;
        break;
      }
    }
  },
};
</script>
