<template>
  <v-flex v-if="schema == null" text-center>
    <h3 style="color: #c9c9c9; margin-top: 20px; margin-bottom: 20px">
      Loading...
    </h3>
  </v-flex>
  <v-flex v-else-if="Object.keys(schema).length == 0" text-center>
    <h3 style="margin-top: 20px; margin-bottom: 20px">No Settings Available</h3>
    <p>The installed plugins do not define any user settings.</p>
  </v-flex>
  <v-list
    v-else
    flat
    style="background: none; padding-top: 0px; margin-top: -5px"
    dense
    expand
  >
    <v-list-group
      color="secondary lighten-2"
      v-for="item in categories"
      :key="item.key"
      :ripple="false"
      value="true"
      no-action
    >
      <template v-slot:activator>
        <v-list-item-content>
          <v-list-item-title>
            {{ item.title }}
          </v-list-item-title>
        </v-list-item-content>
      </template>
      <plugin-settings
        :schema="item.schema"
        :value="item.value"
        :plugin="item.key"
      />
    </v-list-group>
  </v-list>
</template>
<script>
import PluginSettings from "./pluginsettings.vue";
export default {
  components: { PluginSettings },
  computed: {
    schema() {
      return this.$store.state.heedy.plugin_settings_schema;
    },
    categories() {
      let s = this.schema;
      let res = [];

      // Always start with the heedy object, and then go in alphabetical order
      if (s.hasOwnProperty("heedy")) {
        res.push({
          key: "heedy",
          title: "HEEDY",
          schema: s["heedy"],
          value: this.$store.state.app.info.settings["heedy"] || {},
        });
      }

      Object.keys(s)
        .sort()
        .forEach((k) => {
          if (k != "heedy") {
            res.push({
              key: k,
              title: k.toUpperCase(),
              schema: s[k],
              value: this.$store.state.app.info.settings[k] || {},
            });
          }
        });
      return res;
    },
  },
  created() {
    this.$store.dispatch("ReadUserPluginSettingsSchema");
  },
};
</script>