<template>
  <v-form @submit="insert" v-model="formValid">
    <div ref="jsform" v-if="!loading">
      <v-jsonschema-form
        :schema="schema"
        :options="options"
        :model="modified"
      />
    </div>
    <!-- https://github.com/koumoul-dev/vuetify-jsonschema-form/issues/21 -->
    <div
      v-else
      :style="{ height, textAlign: 'center', display: 'flex', margin: 'auto' }"
    >
      <h4 style="margin: auto">Inserting...</h4>
    </div>
    <v-btn dark color="info" type="submit" :loading="loading">Insert</v-btn>
  </v-form>
</template>
<script>
import VJsonschemaForm from "../../dist/vuetify-jsonschema-form.mjs";
import moment from "../../dist/moment.mjs";
export default {
  components: {
    VJsonschemaForm
  },
  props: {
    data: Object
  },
  data: () => ({
    formValid: false,
    loading: false,
    height: "20px",
    options: {
      debug: false,
      disableAll: false,
      autoFoldObjects: false
    },
    modified: {}
  }),
  computed: {
    schema() {
      return {
        type: "object",
        properties: {
          data: {
            title: "Insert Datapoint",
            ...this.data.schema
          }
        },
        required: ["data"]
      };
    }
  },
  methods: {
    insert: async function(event) {
      event.preventDefault();
      if (this.loading) return;
      if (!this.formValid) {
        return;
      }

      this.height = this.$refs.jsform.clientHeight + "px";

      this.loading = true;

      console.log("Inserting datapoint:", this.modified.data);
      let res = await this.$frontend.api(
        "POST",
        `api/objects/${this.data.id}/timeseries`,
        [{ t: moment().unix(), d: this.modified.data }]
      );

      if (!res.response.ok) {
        console.error(res);
        this.loading = false;
        return;
      }
      this.modified = { data: null };
      this.loading = false;
    }
  }
};
</script>
