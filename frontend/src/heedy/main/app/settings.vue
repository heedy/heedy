<template>
  <h-card-page :title="'Update ' + app.name + ' Settings'" :alert="alert">
    <v-form @submit="update" v-model="formValid">
      <v-container fluid grid-list-md>
        <v-layout row>
          <v-flex>
            <v-jsonschema-form
              :schema="schema"
              :model="modified"
              :options="options"
              @error="show"
              @change="show"
              @input="show"
            />
          </v-flex>
        </v-layout>
      </v-container>

      <v-card-actions>
        <v-spacer></v-spacer>

        <v-btn dark color="info" type="submit" :loading="loading">Update Settings</v-btn>
      </v-card-actions>
    </v-form>
  </h-card-page>
</template>
<script>
import VJsonschemaForm from "../../../dist/vuetify-jsonschema-form.mjs";
export default {
  components: {
    VJsonschemaForm
  },
  props: {
    app: Object
  },
  data: function() {
    return {
      formValid: false,
      modified: { ...this.app.settings },
      loading: false,
      alert: "",
      options: {
        debug: false,
        disableAll: false,
        autoFoldObjects: false
      }
    };
  },
  computed: {
    schema() {
      if (this.app.settings_schema.type !== undefined) {
        return this.app.settings_schema;
      }
      let s = {
        type: "object",
        properties: {
          ...this.app.settings_schema
        }
      };
      if (s.properties.required !== undefined) {
        s.required = s.properties.required;
        delete s.properties.required;
      }
      if (s.properties.title !== undefined) {
        s.title = s.properties.title;
        delete s.properties.title;
      }
      if (s.properties.description !== undefined) {
        s.description = s.properties.description;
        delete s.properties.description;
      }
      console.log(s);
      return s;
    }
  },
  methods: {
    show(e) {
      console.log(e);
    },
    update: async function(event) {
      event.preventDefault();
      if (this.loading) return;
      if (!this.formValid) {
        return;
      }

      this.loading = true;
      this.alert = "";

      let modified = {
        settings: {
          ...this.modified
        }
      };
      console.log("Update app settings", this.app.id);

      if (Object.keys(this.modified).length > 0) {
        let result = await this.$app.api(
          "PATCH",
          `api/heedy/v1/apps/${this.app.id}`,
          modified
        );

        if (!result.response.ok) {
          this.alert = result.data.error_description;
          this.loading = false;
          return;
        }

        this.$store.dispatch("readApp", {
          id: this.app.id
        });
      }

      this.loading = false;
      this.$router.push({ path: `/apps/${this.app.id}` });
    },
    del: async function() {}
  }
};
</script>