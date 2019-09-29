<template>
  <div>
    <h-loading v-if="source==null"></h-loading>
    <router-view v-else :source="source"></router-view>
  </div>
</template>
<script>
export default {
  data: () => ({}),
  props: {
    sourceid: String
  },
  watch: {
    sourceid(newValue) {
      this.$store.dispatch("readSource", { id: newValue });
    }
  },
  computed: {
    source() {
      return this.$store.state.heedy.sources[this.sourceid] || null;
    }
  },
  created() {
    this.$store.dispatch("readSource", { id: this.sourceid });
  }
};
</script>