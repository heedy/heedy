<template>
  <h-card-page title="Create a new App" :alert="alert">
    <v-container fluid grid-list-md>
      <v-layout row>
        <v-flex sm5 md4 xs12>
          <h-icon-editor ref="iconEditor" image="settings_input_component"></h-icon-editor>
        </v-flex>
        <v-flex sm7 md8 xs12>
          <v-container>
            <v-text-field label="Name" placeholder="My App" v-model="name"></v-text-field>
            <v-text-field
              label="Description"
              placeholder="This app does stuff"
              v-model="description"
            ></v-text-field>
            <h-scope-editor v-model="scope"></h-scope-editor>
          </v-container>
        </v-flex>
      </v-layout>
    </v-container>

    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn dark color="blue" @click="create" :loading="loading">Create</v-btn>
    </v-card-actions>
  </h-card-page>
</template>
<script>
import api from "../../../rest.mjs";
export default {
  data: () => ({
    description: "",
    scope: "self.objects",
    name: "",
    loading: false,
    alert: ""
  }),
  methods: {
    create: async function() {
      if (this.loading) return;

      this.loading = true;
      this.alert = "";

      let query = {
        name: this.name.trim(),
        description: this.description.trim(),
        scope: this.scope,
        icon: this.$refs.iconEditor.getImage()
      };

      if (query.name.length == 0) {
        this.alert = "A name is required";
        this.loading = false;
        return;
      }

      let result = await api("POST", `api/apps`, query);

      if (!result.response.ok) {
        this.alert = result.data.error_description;
        this.loading = false;
        return;
      }

      // The result comes without the icon, let's set it correctly
      result.data.icon = query.icon;

      this.$store.commit("setApp", result.data);
      this.loading = false;
      this.$router.replace({ path: `/apps/${result.data.id}` });
    }
  }
};
</script>