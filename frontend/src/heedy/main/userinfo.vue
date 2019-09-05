<template>
  <v-card>
    <div style="position:absolute;top:2px;right:16px;" v-if="editable && !editing">
      <v-btn icon @click="editing=true">
        <v-icon style="color:lightgray;opacity:0.3">edit</v-icon>
      </v-btn>
    </div>
    <v-speed-dial bottom right absolute v-if="editable && !editing"
      direction="left"
      v-model="fab"
      transition="slide-x-reverse-transition"
      open-on-hover
    >
        <template v-slot:activator>
        <v-btn
          v-model="fab"
          color="blue darken-2"
          dark
          fab
        >
          <v-icon v-if="fab">close</v-icon>
          <v-icon v-else>add</v-icon>
        </v-btn>
      </template>
      <v-tooltip v-for="item in sourceCreators" :key="item.key" bottom>
        <template v-slot:activator="{ on }">
          <v-btn
            fab
            dark
            small
            color="green"
            v-on="on"
            :to="item.route"
          >
            <v-icon>{{ item.icon }}</v-icon>
          </v-btn>
        </template>
        <span>{{ item.text }}</span>
      </v-tooltip>
      
    </v-speed-dial>
    <v-container grid-list-md>
      <v-layout row wrap>
        <v-flex xs12 sm4 md3 lg2 text-center justify-center>
          <template v-if="!editing">
            <avatar :size="120" :image="user.avatar" :colorHash="user.username" ></avatar>
            <h5 style="color:gray;padding-top:10px">{{user.username}}</h5>
          </template>
          <template v-else>
            <avatar-editor ref="avatarEditor" :image="user.avatar" :colorHash="user.username" ></avatar-editor>
          </template>
        </v-flex>
        <v-flex xs12 sm8 md9 lg10>
          <h2 v-if="!editing">{{ user.name==""?user.username:user.name}}</h2>
          <v-text-field
            v-else
            :label="user.name==''?user.username:user.name"
            solo
            v-model="name"
          ></v-text-field>
          <v-textarea v-if="editing" solo label="No description given." v-model="description"></v-textarea>
          <p v-else-if="user.description!=''">{{ user.description }}</p>
          <p v-else style="color:lightgray;">No description given.</p>
        </v-flex>
      </v-layout>
    </v-container>
    <v-card-actions v-if="editing">
      <v-spacer></v-spacer>
      <v-btn @click="cancel">Cancel</v-btn>
      <v-btn type="submit" color="primary" @click="save" :loading="loading">Save</v-btn>
    </v-card-actions>
  </v-card>
</template>

<script>

import {Avatar, AvatarEditor} from "../components.mjs";

import api from "../api.mjs";

export default {
  components: {
    Avatar,
    AvatarEditor
  },
  data: () => ({
    editing: false,
    modified: {},
    loading: false,
    fab: false
  }),
  props: {
    user: Object,
    editable: Boolean
  },
  methods: {
    cancel() {
      this.loading = false;
      this.editing = false;
      this.modified = {};
    },
    save: async function() {
      if (this.loading) return;
      this.loading = true;
      if (this.$refs.avatarEditor.hasImage()) {
        // We are in the image picker, and an image was chosen
        this.modified.avatar = this.$refs.avatarEditor.getImage();
      }
      console.log(this.modified);
      let result = await api(
        "PATCH",
        `api/heedy/v1/users/${this.user.username}`,
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
        }
      });
    }
  },
  computed: {
    description: {
      get() {
        return this.modified.description || this.user.description;
      },
      set(v) {
        this.modified.description = v;
      }
    },
    name: {
      get() {
        return this.modified["name"] || this.user.name;
      },
      set(v) {
        this.modified.name = v;
      }
    },
    sourceCreators() {
      return this.$store.state.heedy.sourceCreators;
    }
  }
};
</script>