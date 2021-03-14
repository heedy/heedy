<template>
  <v-form @submit="insert" v-model="formValid">
    <div ref="jsform" v-if="!loading">
      <v-jsf :schema="schema" :options="options" v-model="modified">
        <template
          v-for="ins in inserters"
          :slot="`custom-` + ins.k"
          slot-scope="{ value, label, on }"
        >
          <component
            :key="ins.k"
            :is="ins.v"
            :value="value"
            v-on="on"
            :label="label"
          />
        </template>
      </v-jsf>
    </div>
    <!-- https://github.com/koumoul-dev/vuetify-jsonschema-form/issues/21 -->
    <div
      v-else
      :style="{ height, textAlign: 'center', display: 'flex', margin: 'auto' }"
    >
      <h4 style="margin: auto">Inserting...</h4>
    </div>
    <div class="text-center" style="width: 100%">
      <v-btn dark color="info" type="submit" :loading="loading">Insert</v-btn>
    </div>
  </v-form>
</template>
<script>
import moment from "../../../dist/moment.mjs";
export default {
  props: {
    object: Object,
  },
  data: () => ({
    formValid: false,
    loading: false,
    height: "20px",
    options: {},
    modified: {},
  }),
  computed: {
    inserters() {
      let d = this.$store.state.timeseries.customInserters;
      return Object.keys(d).map((k) => ({ k: k, v: d[k] }));
    },
    schema() {
      if (
        this.object.meta.schema.type !== undefined &&
        this.object.meta.schema.type == "object"
      ) {
        return {
          type: "object",
          properties: {
            data: this.object.meta.schema,
          },
          required: ["data"],
        };
      }
      return {
        type: "object",
        properties: {
          data: {
            title: " ",
            ...this.object.meta.schema,
          },
        },
        required: ["data"],
      };
    },
  },
  methods: {
    insert: async function (event) {
      event.preventDefault();
      if (this.loading) return;
      if (!this.formValid) {
        return;
      }

      this.height = this.$refs.jsform.clientHeight + "px";

      this.loading = true;

      console.vlog("Inserting datapoint:", this.modified.data);
      let res = await this.$frontend.rest(
        "POST",
        `api/objects/${encodeURIComponent(this.object.id)}/timeseries`,
        [{ t: moment().unix(), d: this.modified.data }]
      );

      if (!res.response.ok) {
        console.error(res);
        this.loading = false;
        return;
      }
      this.modified = { data: null };
      this.loading = false;
    },
  },
};
</script>
