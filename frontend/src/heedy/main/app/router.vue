<template>
  <div>
    <h-loading v-if="loading"></h-loading>
    <h-not-found v-else-if="app==null" />
    <router-view v-else :app="app"></router-view>
  </div>
</template>
<script>
export default {
  data: () => ({
    loading: true
  }),
  props: {
    appid: String
  },
  watch: {
    appid(newValue) {
      this.loading = true;
      this.$store.dispatch("readApp", {
        id: newValue,
        callback: () => (this.loading = false)
      });
    },
    websocket(nv) {
      if (nv) {
        // If the websocket gets re-connected, re-read the app
        this.$store.dispatch("readApp", {
          id: this.appid
        });
      }
    }
  },
  computed: {
    app() {
      return this.$store.state.heedy.apps[this.appid] || null;
    },
    websocket() {
      return this.$store.state.app.websocket!=null;
    }
  },
  created() {
    this.$store.dispatch("readApp", {
      id: this.appid,
      callback: () => (this.loading = false)
    });
  }
};
</script>