<template>
  <v-tooltip v-if="!isList" bottom>
    <template #activator="{ on }">
      <v-btn
        icon
        v-on="on"
        color="blue darken-2"
        :to="`/apps/${appid}/settings`"
      >
        <v-icon>fas fa-cog</v-icon>
      </v-btn>
    </template>
    <span>App Settings</span>
  </v-tooltip>
  <v-list-item v-else :to="`/apps/${appid}/settings`">
    <v-list-item-icon>
      <v-icon>fas fa-cog</v-icon>
    </v-list-item-icon>
    <v-list-item-content>
      <v-list-item-title>App Settings</v-list-item-title>
    </v-list-item-content>
  </v-list-item>
</template>
<script>
export default {
  props: {
    appid: String,
    isList: Boolean,
  },
};
</script>
