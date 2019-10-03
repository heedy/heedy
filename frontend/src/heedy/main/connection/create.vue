<template>
  <h-card-page title="Create a new Connection" :alert="alert">
    <v-container fluid grid-list-md>
      <v-layout row>
        <v-flex sm5 md4 xs12>
          <h-avatar-editor ref="avatarEditor" image="settings_input_component"></h-avatar-editor>
        </v-flex>
        <v-flex sm7 md8 xs12>
          <v-container>
            <v-text-field label="Name" placeholder="My Connection" v-model="name"></v-text-field>
            <v-text-field
              label="Description"
              placeholder="This connection does stuff"
              v-model="description"
            ></v-text-field>
            <h-scope-editor v-model="scopes"></h-scope-editor>
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
import api from "../../../api.mjs";
export default {
  data: () => ({
    description: "",
    scopes: "",
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
        scopes: this.scopes,
        avatar: this.$refs.avatarEditor.getImage()
      };

      if (query.name.length == 0) {
        this.alert = "A name is required";
        this.loading = false;
        return;
      }

      let result = await api("POST", `api/heedy/v1/connections`, query);

      if (!result.response.ok) {
        this.alert = result.data.error_description;
        this.loading = false;
        return;
      }

      // The result comes without the avatar, let's set it correctly
      result.data.avatar = query.avatar;

      this.$store.commit("setConnection", result.data);
      this.loading = false;
      this.$router.push({ path: `/connections/${result.data.id}` });
    }
  }
};
</script>