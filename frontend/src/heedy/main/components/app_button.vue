<template>
  <v-tooltip v-if="appid != null" bottom>
    <template #activator="{ on }">
      <v-btn icon :to="`/apps/${appid}`" v-on="on">
        <h-icon
          :image="app.icon"
          defaultIcon="settings_input_component"
          :colorHash="appid"
          :size="size"
        ></h-icon
      ></v-btn>
    </template>
    <span>{{ app.name }}</span>
  </v-tooltip>
</template>
<script>
import HIcon from "./icon.vue";
export default {
  components: {
    HIcon,
  },
  props: {
    appid: String || null,
    size: {
      type: Number,
      default: 20,
    },
  },
  watch: {
    appid(newid, oid) {
      if (newid != null && newid != oid) {
        this.$store.dispatch("readApp", {
          id: newid,
        });
      }
    },
  },
  computed: {
    app() {
      let empty_app = {
        id: this.appid,
        icon: "settings_input_component",
        name: "Go to app",
      };
      if (this.$store.state.heedy.apps == null) {
        return empty_app;
      }
      return this.$store.state.heedy.apps[this.appid] || empty_app;
    },
  },
  created() {
    if (this.appid != null) {
      this.$store.dispatch("readApp", {
        id: this.appid,
      });
    }
  },
};
</script>