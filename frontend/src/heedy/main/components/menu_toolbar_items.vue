<template>
  <v-toolbar-items>
    <slot></slot>
    <template v-for="(item, idx) in items.toolbarItems">
      <v-tooltip bottom :key="idx" v-if="!item.toolbar_component">
        <template #activator="{ on }">
          <v-btn icon v-on="on" :to="item.to">
            <v-icon>{{ item.icon }}</v-icon>
          </v-btn>
        </template>
        <span>{{ item.text }}</span>
      </v-tooltip>
      <component
        :key="idx"
        v-else
        :is="item.toolbar_component"
        v-bind="item.toolbar_props || {}"
      />
    </template>
    <v-menu
      offset-y
      v-if="items.toolbarMenu.length > 0 || items.menuItems.length > 0"
    >
      <template v-slot:activator="{ on: menu, attrs }">
        <v-tooltip bottom>
          <template v-slot:activator="{ on: tooltip }">
            <v-btn icon v-bind="attrs" v-on="{ ...tooltip, ...menu }">
              <v-icon>more_vert</v-icon>
            </v-btn>
          </template>
          <span>Menu</span>
        </v-tooltip>
      </template>
      <v-list>
        <template v-for="(item, idx) in items.toolbarMenu">
          <v-list-item
            :key="`toolbar-${idx}`"
            :to="item.to"
            v-if="!item.toolbar_component"
          >
            <v-list-item-icon>
              <v-icon>{{ item.icon }}</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>{{ item.text }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <component
            :key="idx"
            v-else
            :is="item.menu_component"
            v-bind="item.menu_props || {}"
          />
        </template>
        <template v-for="(item, idx) in items.menuItems">
          <v-list-item
            v-if="!item.toolbar_component"
            :key="`menu-${idx}`"
            :to="item.to"
          >
            <v-list-item-icon>
              <v-icon>{{ item.icon }}</v-icon>
            </v-list-item-icon>
            <v-list-item-content>
              <v-list-item-title>{{ item.text }}</v-list-item-title>
            </v-list-item-content>
          </v-list-item>
          <component
            :key="idx"
            v-else
            :is="item.menu_component"
            v-bind="item.menu_props || {}"
          />
        </template>
      </v-list>
    </v-menu>
  </v-toolbar-items>
</template>
<script>
export default {
  props: {
    toolbar: Array,
    maxSize: {
      type: Number,
      default: 1,
    },
  },
  computed: {
    items() {
      const sorted = [...this.toolbar].sort((a, b) => {
        let aWeight = a.weight || 0;
        let bWeight = b.weight || 0;
        return aWeight - bWeight;
      });

      let toolbarItems = [];
      let toolbarMenu = [];
      let menuItems = [];

      for (let i = 0; i < sorted.length; i++) {
        if (sorted[i].toolbar !== undefined && sorted[i].toolbar) {
          if (toolbarItems.length < this.maxSize - 1) {
            toolbarItems.push(sorted[i]);
          } else {
            toolbarMenu.push(sorted[i]);
          }
        } else {
          menuItems.push(sorted[i]);
        }
      }

      if (menuItems.length == 0 && toolbarMenu.length == 1) {
        // There is no need for a menu at all - it is just a toolbar.
        toolbarItems.push(toolbarMenu.pop());
      }

      // Decide which elements to show on the outside, and which elements to show on the inside
      return { toolbarItems, toolbarMenu, menuItems };
    },
  },
};
</script>
