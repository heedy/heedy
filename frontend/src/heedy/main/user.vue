<template >
  <div>
    <loading v-if="user==null"></loading>
    <v-content v-else>
      <v-container fill-height>
        <v-layout>
          <v-flex>
            <user :user="user" editable></user>
          </v-flex>
        </v-layout>
      </v-container>
    </v-content>
  </div>
</template>

<script>
import {Loading} from "../components.mjs";
import User from "./user2.vue";
export default {
  components: {
    User,
    Loading
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
    user: function() {
      let usr = this.$store.state.heedy.users[this.username] || null;
      return usr;
    }
  },
  created() {
    this.$store.dispatch("readUser", { name: this.username });
  }
};
</script>

<style>
</style>