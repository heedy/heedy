<template>
  <v-flex>
    <v-card>
      <div style="position:absolute;top:2px;right:16px;" v-if="editable && !editing">
        <v-btn icon @click="editing = true">
          <v-icon style="color:lightgray;opacity:0.3">edit</v-icon>
        </v-btn>
      </div>
      <v-btn
        v-if="!editing"
        color="blue darken-2"
        dark
        fab
        style="position:absolute;bottom:15px;right:15px;"
        @click.stop="dialog = true"
      >
        <v-icon>add</v-icon>
      </v-btn>

      <v-container grid-list-md fluid>
        <v-layout row wrap>
          <v-flex xs12 sm4 md3 lg2 text-center justify-center>
            <template v-if="!editing">
              <h-icon
                :size="120"
                :image="user.icon"
                defaultIcon="person"
                :colorHash="user.username"
              ></h-icon>
              <h5 style="color:gray;padding-top:10px">{{ user.username }}</h5>
            </template>
            <template v-else>
              <h-icon-editor ref="iconEditor" :image="user.icon" :colorHash="user.username"></h-icon-editor>
            </template>
          </v-flex>
          <v-flex xs12 sm8 md9 lg10>
            <h2 v-if="!editing">{{ user.name == "" ? user.username : user.name }}</h2>
            <v-text-field
              v-else
              :label="user.name == '' ? user.username : user.name"
              solo
              v-model="name"
            ></v-text-field>
            <v-textarea v-if="editing" solo label="No description given." v-model="description"></v-textarea>
            <p v-else-if="user.description != ''">{{ user.description }}</p>
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
    <v-dialog v-model="dialog" max-width="1024">
      <v-card>
        <v-card-title>
          <v-list-item two-line>
            <v-list-item-content>
              <v-list-item-title class="headline mb-1">Add</v-list-item-title>
              <v-list-item-subtitle>
                Create objects that you will manually
                control.
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-card-title>

        <v-card-text>
          <v-row no-gutters>
            <v-col
              v-for="c in objectCreators"
              :key="c.key"
              cols="12"
              xs="12"
              sm="6"
              md="6"
              lg="4"
              xl="3"
            >
              <v-card class="pa-2" outlined tile>
                <v-list-item two-line subheader :to="c.route">
                  <v-list-item-avatar>
                    <h-icon :image="c.icon" :colorHash="c.key" defaultIcon="insert_drive_file"></h-icon>
                  </v-list-item-avatar>
                  <v-list-item-content>
                    <v-list-item-title>{{ c.title }}</v-list-item-title>
                    <v-list-item-subtitle>
                      {{
                      c.description
                      }}
                    </v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
              </v-card>
            </v-col>
          </v-row>
        </v-card-text>
        <v-divider></v-divider>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="secondary" text @click="dialog = false">Cancel</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-flex>
</template>

<script>
import api from "../../../rest.mjs";

export default {
  data: () => ({
    editing: false,
    modified: {},
    loading: false,
    fab: false,
    dialog: false
  }),
  props: {
    user: Object
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
      if (this.$refs.iconEditor.hasImage()) {
        // We are in the image picker, and an image was chosen
        this.modified.icon = this.$refs.iconEditor.getImage();
      }
      console.log(this.modified);
      let result = await api(
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
    editable() {
      if (this.$store.state.app.info.user == null) {
        return false;
      }
      return this.user.username == this.$store.state.app.info.user.username;
    },
    objectCreators() {
      return this.$store.state.heedy.objectCreators;
    }
  }
};
</script>
