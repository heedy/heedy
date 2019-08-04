<template>
    <v-card>
    <div style="position:absolute;top:2px;right:16px;" v-if="editable && !editing">
      <v-btn icon @click="editing=true">
        <v-icon style="color:lightgray;opacity:0.3">edit</v-icon>
      </v-btn>
    </div>
    <v-container grid-list-md>
      <v-layout row wrap>
        <v-flex xs12 sm4 md3 lg2 text-center justify-center>
          <template v-if="!editing">
            <avatar :size="120" :image="stream.avatar" :colorHash="stream.id" ></avatar>
            <h5 style="color:gray;padding-top:10px">{{stream.name}}</h5>
          </template>
          <template v-else>
            <avatar-editor ref="avatarEditor" :image="stream.avatar" :colorHash="stream.id" ></avatar-editor>
          </template>
        </v-flex>
        <v-flex xs12 sm8 md9 lg10>
          <h2 v-if="!editing">{{ stream.fullname==""?stream.name:stream.fullname}}</h2>
          <v-text-field
            v-else
            :label="stream.fullname==''?stream.name:stream.fullname"
            solo
            v-model="fullname"
          ></v-text-field>
          <v-textarea v-if="editing" solo label="No description given." v-model="description"></v-textarea>
          <p v-else-if="stream.description!=''">{{ stream.description }}</p>
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

import {Avatar, AvatarEditor} from "../../heedy/components.mjs";

import api from "../../heedy/api.mjs";

export default {
    components: {
        Avatar,
        AvatarEditor
    },
    data: () => ({
        editing: false,
        modified: {},
        loading: false
    }),
    props: {
        stream: Object,
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
        `api/heedy/v1/source/${this.stream.id}`,
        this.modified
      );
      if (!result.response.ok) {
        this.$store.dispatch("errnotify", result.data);
        this.loading = false;
        return;
      }
      this.$store.dispatch("readSource", {
        id: this.stream.id,
        callback: () => {
          this.cancel();
        }
      });
    }
  },
  computed: {
    description: {
      get() {
        return this.modified.description || this.stream.description;
      },
      set(v) {
        this.modified.description = v;
      }
    },
    fullname: {
      get() {
        return this.modified["fullname"] || this.stream.fullname;
      },
      set(v) {
        this.modified.fullname = v;
      }
    }
  }
}
</script>