<template>
  <div>
    <h-loading v-if="loading"></h-loading>
    <h-not-found v-else-if="source==null" />
    <router-view v-else :source="source"></router-view>
  </div>
</template>
<script>
export default {
  data: () => ({
    loading: true
  }),
  props: {
    sourceid: String
  },
  watch: {
    sourceid(newValue) {
      this.loading = true;
      this.$store.dispatch("readSource", {
        id: newValue,
        callback: () => (this.loading = false)
      });
    }
  },
  computed: {
    source() {
      return this.$store.state.heedy.sources[this.sourceid] || null;
    }
  },
  created() {
    this.$store.dispatch("readSource", {
      id: this.sourceid,
      callback: () => (this.loading = false)
    });
  }
};
</script>