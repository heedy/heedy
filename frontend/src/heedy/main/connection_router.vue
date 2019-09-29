<template>
  <div>
    <h-loading v-if="loading"></h-loading>
    <h-not-found v-else-if="connection==null" />
    <router-view v-else :connection="connection"></router-view>
  </div>
</template>
<script>
export default {
  data: () => ({}),
  props: {
    connectionid: String
  },
  watch: {
    connectionid(newValue) {
      this.$store.dispatch("readConnection", { id: newValue });
    }
  },
  computed: {
    connection() {
      let c = this.$store.state.heedy.connections[this.connectionid] || null;
      return c;
    },
    loading() {
      return this.$store.state.heedy.connections == null;
    }
  },
  created() {
    this.$store.dispatch("readConnection", { id: this.connectionid });
  }
};
</script>