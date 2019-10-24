<template>
  <div>
    <h-loading v-if="loading"></h-loading>
    <h-not-found v-else-if="connection==null" />
    <router-view v-else :connection="connection"></router-view>
  </div>
</template>
<script>
export default {
  data: () => ({
    loading: true
  }),
  props: {
    connectionid: String
  },
  watch: {
    connectionid(newValue) {
      this.loading = true;
      this.$store.dispatch("readConnection", {
        id: newValue,
        callback: () => (this.loading = false)
      });
    }
  },
  computed: {
    connection() {
      return this.$store.state.heedy.connections[this.connectionid] || null;
    }
  },
  created() {
    this.$store.dispatch("readConnection", {
      id: this.connectionid,
      callback: () => (this.loading = false)
    });
  }
};
</script>