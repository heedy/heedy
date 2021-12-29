<template>
  <h-card-page :title="'Update ' + app.name" :alert="alert">
    <v-form @submit="update">
      <v-container fluid grid-list-md>
        <v-layout row>
          <v-flex sm5 md4 xs12>
            <h-icon-editor
              ref="iconEditor"
              :image="app.icon"
              :colorHash="app.id"
              defaultIcon="settings_input_component"
            ></h-icon-editor>
          </v-flex>
          <v-flex sm7 md8 xs12>
            <v-container>
              <v-text-field
                label="Name"
                placeholder="My App"
                autofocus
                v-model="name"
              ></v-text-field>
              <v-text-field
                label="Description"
                placeholder="This app does stuff"
                v-model="description"
              ></v-text-field>
              <h-scope-editor v-model="scope"></h-scope-editor>
              <v-layout row style="padding: 0; margin-top: -25px">
                <v-flex style="padding-right: 0; margin-bottom: -40px">
                  <v-checkbox
                    v-if="
                      app.access_token === undefined || app.access_token != ''
                    "
                    style="
                      margin-top: 0;
                      padding-bottom: 0;
                      padding-top: 0;
                      margin-bottom: 0;
                      padding-right: 0;
                    "
                    v-model="reset_token"
                    label="Reset Token"
                  ></v-checkbox>
                </v-flex>
                <v-flex
                  style="
                    text-align: right;
                    padding-left: 0;
                    margin-bottom: -40px;
                  "
                >
                  <v-checkbox
                    style="
                      margin-top: 0;
                      padding-bottom: 0;
                      padding-top: 0;
                      margin-bottom: 0;
                      padding-right: 0;
                      float: right;
                    "
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
        <v-btn dark color="red" @click="del" :loading="loading">Delete</v-btn>
        <v-spacer></v-spacer>

        <v-btn color="primary" type="submit" :loading="loading">Update</v-btn>
      </v-card-actions>
    </v-form>
  </h-card-page>
</template>
<script>
export default {
  props: {
    app: Object,
  },
  data: () => ({
    modified: {},
    reset_token: false,
    loading: false,
    alert: "",
  }),
  methods: {
    update: async function (e) {
      e.preventDefault();
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

      console.vlog("Update app", this.app.id, {
        ...this.modified,
      });

      if (Object.keys(this.modified).length > 0) {
        let result = await this.$frontend.rest(
          "PATCH",
          `api/apps/${encodeURIComponent(this.app.id)}`,
          this.modified
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
      this.$router.go(-1);
    },
    del: async function () {
      if (
        confirm(
          `Are you sure you want to delete '${this.app.name}'? You can disable it instead, which will keep any data this app has gathered.`
        )
      ) {
        let res = await this.$frontend.rest(
          "DELETE",
          `/api/apps/${encodeURIComponent(this.app.id)}`
        );
        if (!res.response.ok) {
          this.alert = res.data.error_description;
        } else {
          this.alert = "";
          this.$router.push("/apps");
        }
      }
    },
  },
  computed: {
    description: {
      get() {
        if (this.modified.description !== undefined) {
          return this.modified.description;
        }
        return this.app.description;
      },
      set(v) {
        this.$frontend.vue.set(this.modified, "description", v);
      },
    },
    name: {
      get() {
        if (this.modified.name !== undefined) {
          return this.modified.name;
        }
        return this.app.name;
      },
      set(v) {
        this.$frontend.vue.set(this.modified, "name", v);
      },
    },
    scope: {
      get() {
        if (this.modified.scope !== undefined) {
          return this.modified.scope;
        }
        return this.app.scope;
      },
      set(v) {
        this.$frontend.vue.set(this.modified, "scope", v);
      },
    },
    enabled: {
      get() {
        if (this.modified["enabled"] === undefined) {
          return this.app.enabled;
        }
        return this.modified["enabled"];
      },
      set(v) {
        this.$frontend.vue.set(this.modified, "enabled", v);
      },
    },
  },
};
</script>