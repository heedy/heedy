<template>
  <div>
    <v-form @submit="save">
      <v-container grid-list-md fluid>
        <v-layout row wrap>
          <v-flex xs12 sm4 md3 lg2 text-center justify-center>
            <h-icon-editor
              ref="iconEditor"
              :image="user.icon"
              :colorHash="user.username"
            ></h-icon-editor>
          </v-flex>
          <v-flex xs12 sm8 md9 lg10>
            <v-text-field
              :label="user.name == '' ? user.username : user.name"
              solo
              v-model="name"
            ></v-text-field>
            <v-textarea
              solo
              label="Add a short description here"
              v-model="description"
            ></v-textarea>
          </v-flex>
        </v-layout>
        <v-layout row wrap style="border: 1px">
          <v-flex xs12 sm6 md6 lg6>
            <v-text-field
              type="password"
              filled
              clearable
              label="New Password (Optional)"
              v-model="password"
            ></v-text-field>
          </v-flex>
          <v-flex xs12 sm6 md6 lg6>
            <v-text-field
              type="password"
              filled
              clearable
              label="Repeat Password"
              v-model="password2"
            ></v-text-field>
          </v-flex>
        </v-layout>
      </v-container>
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn type="submit" color="primary" :loading="loading">Save</v-btn>
      </v-card-actions>
    </v-form>
  </div>
</template>
<script>
export default {
  data: () => ({
    loading: false,
    modified: {},
    password: "",
    password2: "",
  }),
  computed: {
    user() {
      return this.$store.state.app.info.user;
    },
    description: {
      get() {
        if (this.modified.description === undefined) {
          return this.user.description;
        }
        return this.modified.description;
      },
      set(v) {
        this.modified.description = v;
      },
    },
    name: {
      get() {
        if (this.modified.name === undefined) {
          return this.user.name;
        }
        return this.modified["name"];
      },
      set(v) {
        this.modified.name = v;
      },
    },
  },
  methods: {
    cancel() {
      this.loading = false;
      this.modified = {};
    },
    save: async function (e) {
      e.preventDefault();

      if (this.loading) return;
      if (this.password != "" || this.password2 != "") {
        if (this.password != this.password2) {
          alert("Passwords do not match.");
          return;
        }
        this.modified.password = this.password;
      }
      this.loading = true;
      if (this.$refs.iconEditor.hasImage()) {
        // We are in the image picker, and an image was chosen
        this.modified.icon = this.$refs.iconEditor.getImage();
      }
      console.vlog(this.modified);
      let result = await this.$frontend.rest(
        "PATCH",
        `api/users/${this.user.username}`,
        this.modified
      );
      if (!result.response.ok) {
        this.$store.dispatch("errnotify", result.data);
        this.loading = false;
        return;
      }
      this.$store.dispatch("readUser", {
        username: this.user.username,
        callback: () => {
          this.cancel();
          this.$router.push({ path: `/users/${this.user.username}` });
        },
      });
    },
  },
};
</script>