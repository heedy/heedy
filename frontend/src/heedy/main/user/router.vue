<template>
  <div>
    <h-loading v-if="loading"></h-loading>
    <h-not-found v-else-if="user==null" />
    <router-view v-else :user="user"></router-view>
  </div>
</template>
<script>
export default {
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
    username(newValue) {
      this.$store.dispatch("readUser", { username: newValue });
    }
  },
  computed: {
    user() {
      let c = this.$store.state.heedy.users[this.username] || null;
      return c;
    },
    loading() {
      return this.user == null;
    }
  },
  created() {
    this.$store.dispatch("readUser", { username: this.username });
  }
};
</script>