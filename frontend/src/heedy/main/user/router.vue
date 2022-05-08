<template>
  <div>
    <h-loading v-if="loading"></h-loading>
    <h-not-found v-else-if="user == null" />
    <router-view v-else :user="user"></router-view>
  </div>
</template>
<script>
export default {
  data: () => ({
    loading: true,
  }),
  head() {
    return {
      title: this.username,
    };
  },
  props: {
    username: {
      type: String,
      default: function () {
        if (this.$store.state.app.info.user != null) {
          return this.$store.state.app.info.user.username;
        }
        return "";
      },
    },
  },
  watch: {
    username(newValue) {
      this.loading = true;
      this.$store.dispatch("readUser", {
        username: newValue,
        callback: () => (this.loading = false),
      });
    },
    websocket(nv) {
      if (nv && this.username != this.$store.state.app.info.user?.username) {
        // If the websocket gets re-connected, re-read the user (unless the user is read through appinfo)
        this.$store.dispatch("readUser", {
          username: this.username,
        });
      }
    }
  },
  computed: {
    user() {
      return this.$store.state.heedy.users[this.username] || null;
    },
    websocket() {
      return this.$store.state.app.websocket!=null;
    }
  },
  created() {
    this.$store.dispatch("readUser", {
      username: this.username,
      callback: () => (this.loading = false),
    });
  },
};
</script>