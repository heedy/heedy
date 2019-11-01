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
    }
  },
  computed: {
    app() {
      return this.$store.state.heedy.apps[this.appid] || null;
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