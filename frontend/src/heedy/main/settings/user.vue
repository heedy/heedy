<template>
  <div>
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
    </v-container>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn type="submit" color="primary" @click="save" :loading="loading"
        >Save</v-btn
      >
    </v-card-actions>
  </div>
</template>
<script>
export default {
  data: () => ({
    loading: false,
    modified: {},
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
    save: async function () {
      if (this.loading) return;
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