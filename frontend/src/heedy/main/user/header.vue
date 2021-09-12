<template>
  <v-flex>
    <v-card>
      <div
        style="position: absolute; top: 2px; right: 2px; background: none"
        v-if="toolbar.length > 0"
      >
        <v-toolbar dense collapse flat>
          <h-menu-toolbar-items :toolbar="toolbar" :maxSize="toolbarSize" />
        </v-toolbar>
      </div>
      <v-btn
        color="blue darken-2"
        dark
        fab
        style="position: absolute; bottom: 15px; right: 15px"
        @click.stop="dialog = true"
      >
        <v-icon>add</v-icon>
      </v-btn>

      <v-container grid-list-md fluid>
        <v-layout row wrap>
          <v-flex xs12 sm4 md3 lg2 text-center justify-center>
            <h-icon
              :size="120"
              :image="user.icon"
              defaultIcon="person"
              :colorHash="user.username"
            ></h-icon>
            <h5 style="color: gray; padding-top: 10px">
              {{ user.username }}
            </h5>
          </v-flex>
          <v-flex xs12 sm8 md9 lg10>
            <h2>
              {{ user.name == "" ? user.username : user.name }}
            </h2>
            <p v-if="user.description != ''">{{ user.description }}</p>
            <p v-else style="color: lightgray"></p>
          </v-flex>
        </v-layout>
      </v-container>
    </v-card>
    <v-dialog v-model="dialog" max-width="1024">
      <v-card>
        <v-card-title>
          <v-list-item two-line>
            <v-list-item-content>
              <v-list-item-title class="headline mb-1">Add</v-list-item-title>
              <v-list-item-subtitle>
                Create objects that you will manually control.
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
                <v-list-item two-line subheader @click="() => runCreator(c)">
                  <v-list-item-avatar>
                    <h-icon
                      :image="c.icon"
                      :colorHash="c.key"
                      defaultIcon="insert_drive_file"
                    ></h-icon>
                  </v-list-item-avatar>
                  <v-list-item-content>
                    <v-list-item-title>{{ c.title }}</v-list-item-title>
                    <v-list-item-subtitle>
                      {{ c.description }}
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
import api from "../../../util.mjs";

export default {
  data: () => ({
    fab: false,
    dialog: false,
  }),
  props: {
    user: Object,
  },
  methods: {
    runCreator(c) {
      if (c.route !== undefined) {
        this.$router.push({ path: c.route });
      } else {
        // There must be a function to call
        c.fn();
      }
    },
  },
  computed: {
    objectCreators() {
      return this.$store.state.heedy.objectCreators;
    },
    toolbar() {
      return Object.values(
        this.$store.state.heedy.userMenu.reduce(
          (o, m) => ({ ...o, ...m(this.user) }),
          {}
        )
      );
    },
    toolbarSize() {
      if (this.$vuetify.breakpoint.xs) {
        return 1;
      }
      if (this.$vuetify.breakpoint.sm) {
        return 2;
      }
      if (this.$vuetify.breakpoint.md) {
        return 3;
      }
      return 6;
    },
  },
};
</script>
