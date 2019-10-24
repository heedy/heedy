<template>
  <v-flex>
    <h-source-list :sources="sources" />
  </v-flex>
</template>
<script>
export default {
  props: {
    user: Object
  },
  computed: {
    sources() {
      return Object.keys(
        this.$store.state.heedy.userSources[this.user.username] || {}
      ).map(id => this.$store.state.heedy.sources[id]);
    }
  },
  watch: {
    user: function(u) {
      this.$store.dispatch("readUserSources", { username: u.username });
    }
  },

  created() {
    this.$store.dispatch("readUserSources", { username: this.user.username });
  }
};
</script>