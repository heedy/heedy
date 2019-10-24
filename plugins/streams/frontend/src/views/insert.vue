<template>
  <v-form @submit="insert" v-model="formValid">
    <v-jsonschema-form
      :schema="schema"
      :options="options"
      :model="modified"
      @error="show"
      @change="show"
      @input="show"
    />
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
    show(e) {
      console.log(e, this.modified);
    },
    insert: async function(event) {
      event.preventDefault();
      if (this.loading) return;
      if (!this.formValid) {
        return;
      }

      this.loading = true;
      console.log("Inserting datapoint:", this.modified.data);
      let res = await this.$app.api(
        "POST",
        `api/heedy/v1/sources/${this.data.id}/data`,
        [{ t: moment().unix(), d: this.modified.data }]
      );
      this.loading = false;
      if (!res.response.ok) {
        console.error(res);
        return;
      }
      this.modified = { data: null };
    }
  },
  created() {
    console.log("CREATED", this.data);
  }
};
</script>