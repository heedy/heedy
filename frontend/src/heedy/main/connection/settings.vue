<template>
  <h-card-page :title="'Update ' + connection.name + ' Settings'" :alert="alert">
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
export default {
  props: {
    connection: Object
  },
  data: function() {
    return {
      formValid: false,
      modified: { ...this.connection.settings },
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
      if (this.connection.settings_schema.type !== undefined) {
        return this.connection.settings_schema;
      }
      let s = {
        type: "object",
        properties: {
          ...this.connection.settings_schema
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
      console.log("Update connection settings", this.connection.id);

      if (Object.keys(this.modified).length > 0) {
        let result = await this.$app.api(
          "PATCH",
          `api/heedy/v1/connections/${this.connection.id}`,
          modified
        );

        if (!result.response.ok) {
          this.alert = result.data.error_description;
          this.loading = false;
          return;
        }

        this.$store.dispatch("readConnection", {
          id: this.connection.id
        });
      }

      this.loading = false;
      this.$router.push({ path: `/connections/${this.connection.id}` });
    },
    del: async function() {}
  }
};
</script>