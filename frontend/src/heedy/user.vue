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
import Loading from "./loading.mjs";
import User from "./components/user.mjs";
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
        if (this.$store.state.info.user != null) {
          return this.$store.state.info.user.name;
        }
        return "";
      }
    }
  },
  watch: {
    username: function(newValue) {
      console.log("REad USer", newValue);
      this.$store.dispatch("readUser", { name: newValue });
    }
  },
  computed: {
    user: function() {
      console.log("Compued");
      let usr = this.$store.state.users[this.username] || null;
      console.log(this.username, usr, this.$store.state.users);
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