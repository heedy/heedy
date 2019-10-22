<template>
  <h-card-page :title="'Update '+ connection.name" :alert="alert">
    <v-container fluid grid-list-md>
      <v-layout row>
        <v-flex sm5 md4 xs12>
          <h-icon-editor ref="iconEditor" :image="connection.icon" :colorHash="connection.id"></h-icon-editor>
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
            <v-layout row style="padding:0; margin-top:-25px; ">
              <v-flex style="padding-right: 0; margin-bottom: -40px;">
                <v-checkbox
                  style="margin-top: 0; padding-bottom: 0; padding-top:0; margin-bottom:0; padding-right: 0;"
                  v-model="reset_token"
                  label="Reset Token"
                ></v-checkbox>
              </v-flex>
              <v-flex style="text-align: right; padding-left: 0; margin-bottom: -40px;">
                <v-checkbox
                  style="margin-top: 0; padding-bottom: 0; padding-top:0; margin-bottom:0; padding-right: 0;float: right;"
                  v-model="enabled"
                  label="Enabled"
                ></v-checkbox>
              </v-flex>
            </v-layout>
          </v-container>
        </v-flex>
      </v-layout>
    </v-container>

    <v-card-actions>
      <v-btn v-if="!enabled" dark color="red" @click="del" :loading="loading">Delete</v-btn>
      <v-spacer></v-spacer>

      <v-btn dark color="blue" @click="update" :loading="loading">Update</v-btn>
    </v-card-actions>
  </h-card-page>
</template>
<script>
export default {
  props: {
    connection: Object
  },
  data: () => ({
    modified: {},
    reset_token: false,
    loading: false,
    alert: ""
  }),
  methods: {
    update: async function() {
      if (this.loading) return;

      this.loading = true;
      this.alert = "";

      if (this.$refs.iconEditor.hasImage()) {
        // We are in the image picker, and an image was chosen
        this.modified.icon = this.$refs.iconEditor.getImage();
      }
      if (this.reset_token) {
        this.modified.access_token = "reset";
      }

      console.log("Update connection", this.connection.id, {
        ...this.modified
      });

      if (Object.keys(this.modified).length > 0) {
        let result = await this.$app.api(
          "PATCH",
          `api/heedy/v1/connections/${this.connection.id}`,
          this.modified
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
  },
  computed: {
    description: {
      get() {
        return this.modified.description || this.connection.description;
      },
      set(v) {
        this.$app.vue.set(this.modified, "description", v);
      }
    },
    name: {
      get() {
        return this.modified["name"] || this.connection.name;
      },
      set(v) {
        this.$app.vue.set(this.modified, "name", v);
      }
    },
    scopes: {
      get() {
        return this.modified["scopes"] || this.connection.scopes;
      },
      set(v) {
        this.$app.vue.set(this.modified, "scopes", v);
      }
    },
    enabled: {
      get() {
        if (this.modified["enabled"] === undefined) {
          return this.connection.enabled;
        }
        return this.modified["enabled"];
      },
      set(v) {
        this.$app.vue.set(this.modified, "enabled", v);
      }
    }
  }
};
</script>