<template>
  <h-card-page :title="'Update ' + app.name + ' Settings'" :alert="alert">
    <v-form @submit="update" v-model="formValid">
      <v-container fluid grid-list-md>
        <v-layout row>
          <v-flex>
            <v-jsf
              :schema="schema"
              v-model="modified"
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

        <v-btn dark color="info" type="submit" :loading="loading"
          >Update Settings</v-btn
        >
      </v-card-actions>
    </v-form>
  </h-card-page>
</template>
<script>
export default {
  props: {
    app: Object,
  },
  data: function () {
    return {
      formValid: false,
      modified: { ...this.app.settings },
      loading: false,
      alert: "",
      options: {
        debug: false,
        disableAll: false,
        autoFoldObjects: false,
      },
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
          ...this.app.settings_schema,
        },
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
      console.vlog(s);
      return s;
    },
  },
  methods: {
    show(e) {
      console.vlog(e);
    },
    update: async function (event) {
      event.preventDefault();
      if (this.loading) return;
      if (!this.formValid) {
        return;
      }

      this.loading = true;
      this.alert = "";

      let modified = {
        settings: {
          ...this.modified,
        },
      };
      console.vlog("Update app settings", this.app.id);

      if (Object.keys(this.modified).length > 0) {
        let result = await this.$frontend.rest(
          "PATCH",
          `api/apps/${encodeURIComponent(this.app.id)}`,
          modified
        );

        if (!result.response.ok) {
          this.alert = result.data.error_description;
          this.loading = false;
          return;
        }

        this.$store.dispatch("readApp", {
          id: this.app.id,
        });
      }

      this.loading = false;
      this.$router.push({ path: `/apps/${encodeURIComponent(this.app.id)}` });
    },
    del: async function () {},
  },
};
</script>
