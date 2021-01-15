<template>
  <h-object-creator v-model="object" :validator="validate">
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
  </h-object-creator>
</template>
<script>
export default {
  data: () => ({
    object: {
      name: "",
      type: "timeseries",
      meta: {
        schema: { type: "number" },
      },
    },
  }),
  computed: {
    types() {
      return Object.values(this.$store.state.timeseries.types);
    },
    schema: {
      get() {
        return this.object.meta.schema;
      },
      set(s) {
        this.object = {
          ...this.object,
          meta: { ...this.object.meta, schema: s },
        };
      },
    },
    curtype: {
      get() {
        let ct = this.$route.params["datatype"] || "";
        return ct;
      },
      set(v) {
        this.$router.replace(`/create/object/timeseries/${v}`);
      },
    },
  },
  methods: {
    validate(o) {
      if (o.meta.schema == null) {
        return "Invalid Schema";
      }
      return "";
    },
  },
  watch: {
    "$route.params": function (params) {
      let ct = params["datatype"] || "";
      if (
        ct == "" ||
        (this.$store.state.timeseries.types[ct] === undefined && ct != "custom")
      ) {
        this.$router.replace(`/create/object/timeseries/number`);
      } else if (ct != "custom") {
        this.schema = this.$store.state.timeseries.types[ct].schema;
      }
    },
  },
  created() {
    let ct = this.$route.params["datatype"] || "";
    if (
      ct == "" ||
      (this.$store.state.timeseries.types[ct] === undefined && ct != "custom")
    ) {
      this.$router.replace(`/create/object/timeseries/number`);
    } else if (ct != "custom") {
      this.schema = this.$store.state.timeseries.types[ct].schema;
    }
  },
};
</script>
