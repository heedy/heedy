<template >
  <div>
    <vue-headful :title="title"></vue-headful>
    <loading v-if="user==null"></loading>
    <v-content v-else>
      <v-container grid-list-xl>
        <v-layout fill-height column>
          <v-flex>
            <user-info :user="user" :editable="editable"></user-info>
          </v-flex>
          <v-flex>
            <v-card>
              <v-list two-line subheader>
                <v-list-item v-for="item in sources" :key="item.id" :to="`/source/${item.id}`">
                  <v-list-item-avatar>
                    <avatar :image="item.avatar" :colorHash="item.id" ></avatar>
                  </v-list-item-avatar>
                  <v-list-item-content>
                    <v-list-item-title>{{ item.name }}</v-list-item-title>
                    <v-list-item-subtitle>{{ item.description }}</v-list-item-subtitle>
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
          return this.$store.state.app.info.user.username;
        }
        return "";
      }
    }
  },
  watch: {
    username: function(newValue) {
      this.$store.dispatch("readUser", { username: newValue });
      this.$store.dispatch("readUserSources", { username: newValue });
    }
  },
  computed: {
    user() {
      let usr = this.$store.state.heedy.users[this.username] || null;
      return usr;
    },
    sources() {
      return (this.$store.state.heedy.userSources[this.username] || []).map((id)=> this.$store.state.heedy.sources[id]);
    },
    editable() {
      if (this.$store.state.app.info.user == null) {
          return false;
        }
      return this.username == this.$store.state.app.info.user.username;
    },
    title() {
            let u = this.user;
            if (u==null) {
                return "loading... | heedy";
            }
            let n = u.name;
            if (n=="") {
              n = u.username;
            }
            return n + " | heedy";
        }
  },
  created() {
    this.$store.dispatch("readUser", { username: this.username });
    this.$store.dispatch("readUserSources", { username: this.username });
  }
};
</script>

<style>
</style>