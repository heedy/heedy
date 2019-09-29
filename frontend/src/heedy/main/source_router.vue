<template>
  <div>
    <vue-headful :title="title"></vue-headful>
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
    },
    title() {
      let s = this.source;
      if (s == null) {
        return "loading... | heedy";
      }
      return s.name + " | heedy";
    }
  },
  created() {
    this.$store.dispatch("readSource", { id: this.sourceid });
  }
};
</script>