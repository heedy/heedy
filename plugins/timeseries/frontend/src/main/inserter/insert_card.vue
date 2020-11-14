<template>
  <v-card>
    <v-card-title>
      <router-link
        :to="`/objects/${object.id}`"
        style="text-decoration: inherit; color: inherit"
      >
        <h-icon
          :image="object.icon"
          :defaultIcon="defaultIcon"
          :colorHash="object.id"
          :size="30"
        ></h-icon>
        <span style="padding-left: 10px">{{ object.name }}</span>
      </router-link>
      <v-spacer />
      <v-tooltip v-if="object.app != null" bottom>
        <template #activator="{ on }">
          <v-btn icon :to="`/apps/${object.app}`" v-on="on">
            <h-icon
              :image="app.icon"
              :defaultIcon="defaultIcon"
              :colorHash="app.id"
              :size="20"
            ></h-icon
          ></v-btn>
        </template>
        <span>{{ app.name }}</span>
      </v-tooltip>
    </v-card-title>
    <v-card-text>
      <inserter :object="object"></inserter>
    </v-card-text>
  </v-card>
</template>
<script>
import Inserter from "./insert.vue";
export default {
  props: {
    object: Object,
  },
  components: {
    Inserter,
  },
  watch: {
    object(newobj, oobj) {
      if (newobj.app != null && newobj.app != oobj.app) {
        this.$store.dispatch("readApp", {
          id: newobj.app,
        });
      }
    },
  },
  computed: {
    defaultIcon() {
      return (
        this.$store.state.heedy.object_types["timeseries"].icon ||
        "brightness_1"
      );
    },
    app() {
      let empty_app = {
        id: this.object.id,
        icon: "settings_input_component",
        name: "Go to app",
      };
      if (this.$store.state.heedy.apps == null) {
        return empty_app;
      }
      return this.$store.state.heedy.apps[this.object.app] || empty_app;
    },
  },
  created() {
    if (this.object.app != null) {
      this.$store.dispatch("listApps");
    }
  },
};
</script>