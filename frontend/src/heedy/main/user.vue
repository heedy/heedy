<template >
  <div>
    <loading v-if="user==null"></loading>
    <v-content v-else>
      <v-container grid-list-xl>
        <v-layout>
          <v-flex>
            <user-info :user="user" :editable="editable"></user-info>
          </v-flex>
        </v-layout>
        <v-layout>
          <v-flex>
            <v-card>
              <v-list two-line subheader>
                <v-list-item>
                  <v-list-item-avatar>
                    <v-icon>accessibility</v-icon>
                  </v-list-item-avatar>
                  <v-list-item-content>
                    <v-list-item-title>Hi there</v-list-item-title>
                    <v-list-item-subtitle>Hey theere I am a </v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
              </v-list>
            </v-card>
          </v-flex>
        </v-layout>
      </v-container>
    </v-content>
  </div>
</template>

<script>
import {Loading, Avatar} from "../components.mjs";
import UserInfo from "./userinfo.vue";
export default {
  components: {
    UserInfo,
    Loading,
    Avatar
  },
  data: () => ({}),

  props: {
    username: {
      type: String,
      default: function() {
        if (this.$store.state.app.info.user != null) {
          return this.$store.state.app.info.user.name;
        }
        return "";
      }
    }
  },
  watch: {
    username: function(newValue) {
      this.$store.dispatch("readUser", { name: newValue });
    }
  },
  computed: {
    user() {
      let usr = this.$store.state.heedy.users[this.username] || null;
      return usr;
    },
    editable() {
      if (this.$store.state.app.info.user == null) {
          return false;
        }
      return this.username == this.$store.state.app.info.user.name;
    }
  },
  created() {
    this.$store.dispatch("readUser", { name: this.username });
  }
};
</script>

<style>
</style>