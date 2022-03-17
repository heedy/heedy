<template>
  <div>
    <h-loading v-if="loading"></h-loading>
    <h-not-found v-else-if="object==null" />
    <router-view v-else :object="object"></router-view>
  </div>
</template>
<script>
export default {
  data: () => ({
    loading: true
  }),
  props: {
    objectid: String
  },
  watch: {
    objectid(newValue) {
      this.loading = true;
      this.$store.dispatch("readObject", {
        id: newValue,
        callback: () => (this.loading = false)
      });
    },
    websocket(nv) {
      if (nv) {
        // If the websocket gets re-connected, re-read the object
        this.$store.dispatch("readObject", {
          id: this.objectid
        });
      }
    }
  },
  computed: {
    object() {
      return this.$store.state.heedy.objects[this.objectid] || null;
    },
    websocket() {
      return this.$store.state.app.websocket!=null;
    }
  },
  created() {
    this.$store.dispatch("readObject", {
      id: this.objectid,
      callback: () => (this.loading = false)
    });
  }
};
</script>