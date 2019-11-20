<template>
  <v-flex>
    <h-object-list :objects="objects" />
  </v-flex>
</template>
<script>
export default {
  props: {
    user: Object
  },
  computed: {
    objects() {
      return Object.keys(
        this.$store.state.heedy.userObjects[this.user.username] || {}
      ).map(id => this.$store.state.heedy.objects[id]);
    }
  },
  watch: {
    user: function(u) {
      this.$store.dispatch("readUserObjects", { username: u.username });
    }
  },

  created() {
    this.$store.dispatch("readUserObjects", { username: this.user.username });
  }
};
</script>