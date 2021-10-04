<template>
  <v-list
    flat
    style="background: none; padding-top: 0px; margin-top: -5px"
    dense
    expand
  >
    <v-list-group
      color="secondary lighten-2"
      v-for="item in items"
      :key="item.type"
      :ripple="false"
      value="true"
      no-action
    >
      <template v-slot:activator>
        <v-list-item-content>
          <v-list-item-title>
            <v-icon v-if="item.icon !== undefined" style="margin-right: 5px">{{
              item.icon
            }}</v-icon>
            {{ item.list_title }}
          </v-list-item-title>
        </v-list-item-content>
      </template>
      <component
        :is="item.list_component"
        :objects="item.objects"
        :defaultIcon="item.icon"
        :showApps="showApps"
      />
    </v-list-group>
  </v-list>
</template>
<script>
import ListDefault from "./list_default.vue";
export default {
  props: {
    objects: Array,

    showApps: {
      type: Boolean,
      default: true,
    },
  },
  computed: {
    items() {
      let srcobj = this.objects
        .filter((o) => o != null)
        .reduce((o, s) => {
          if (o[s.type] === undefined) {
            o[s.type] = [];
          }
          o[s.type].push(s);
          return o;
        }, {});
      Object.keys(srcobj).forEach((k) =>
        srcobj[k].sort((a, b) => {
          if (a.app == b.app) {
            return a.name.localeCompare(b.name);
          }
          if (a.app == null) {
            return -1;
          }
          if (b.app == null) {
            return 1;
          }
          return a.app.localeCompare(b.app);
        })
      );

      let srcType = this.$store.state.heedy.object_types;

      return Object.keys(srcobj).map((k) => ({
        type: k,
        list_title: k.charAt(0).toUpperCase() + k.substring(1) + "s",
        objects: srcobj[k],
        list_component: ListDefault,
        ...(srcType[k] || {}),
      }));
    },
  },
};
</script>