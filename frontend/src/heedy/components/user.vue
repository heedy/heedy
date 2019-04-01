<template>
  <v-card>
    <div style="position:absolute;bottom:2px;right:2px;" v-if="editable && !editing">
      <v-btn icon @click="editing=true">
        <v-icon style="color:lightgray">edit</v-icon>
      </v-btn>
    </div>
    <v-container grid-list-md>
      <v-layout row wrap>
        <v-flex xs12 sm4 md3 lg2 text-xs-center justify-center>
          <template v-if="!editing">
            <avatar :size="120" :image="user.avatar"></avatar>
            <h5 style="color:gray;padding-top:10px">{{user.name}}</h5>
          </template>
          <template v-else>
            <croppa :width="160" :height="160" v-model="imageCropper"></croppa>
            <v-btn small flat>Icon Chooser</v-btn>
          </template>
        </v-flex>
        <v-flex xs12 sm8 md9 lg10>
          <h2 v-if="!editing">{{ user.fullname==""?user.name:user.fullname}}</h2>
          <v-text-field
            v-else
            :label="user.fullname==''?user.name:user.fullname"
            solo
            v-model="fullname"
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
import Croppa from "vue-croppa";
import "vue-croppa/dist/vue-croppa.css";

import Avatar from "./avatar.mjs";

import { api } from "../../main.mjs";

export default {
  components: {
    Croppa: Croppa.component,
    Avatar
  },
  data: () => ({
    editing: false,
    editingIcon: false,
    modified: {},
    loading: false,
    imageCropper: {}
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
      this.loading = true;
      if (!this.editingIcon && this.imageCropper.hasImage()) {
        // We are in the image picker, and an image was chosen
        this.modified.avatar = this.imageCropper.generateDataUrl();
        console.log(this.modified.avatar);
      }
      let result = await api(
        "POST",
        `api/heedy/v1/user/${this.user.name}`,
        this.modified
      );
      if (!result.response.ok) {
        this.$store.dispatch("errnotify", result.data);
        this.loading = false;
        return;
      }
      this.$store.dispatch("readUser", {
        name: this.user.name,
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
    fullname: {
      get() {
        return this.modified["fullname"] || this.user.fullname;
      },
      set(v) {
        this.modified.fullname = v;
      }
    }
  }
};
</script>