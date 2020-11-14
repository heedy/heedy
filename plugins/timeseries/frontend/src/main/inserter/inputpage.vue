<template>
  <h-page-container>
    <v-list
      flat
      style="background: none; padding-top: 0px; margin-top: -5px"
      dense
      expand
    >
      <v-list-group
        v-if="yourInserters.length > 0"
        color="secondary lighten-2"
        :ripple="false"
        value="true"
        no-action
      >
        <template v-slot:activator>
          <v-list-item-content>
            <v-list-item-title> Your Items </v-list-item-title>
          </v-list-item-content>
        </template>
        <v-row>
          <v-col
            v-for="ts in yourInserters"
            :key="ts.id"
            cols="12"
            sm="12"
            md="6"
            lg="6"
            xl="4"
          >
            <insert-card :object="ts" />
          </v-col>
        </v-row>
      </v-list-group>
      <v-list-group
        v-if="appInserters.length > 0"
        color="secondary lighten-2"
        :ripple="false"
        value="true"
        no-action
      >
        <template v-slot:activator>
          <v-list-item-content>
            <v-list-item-title> Managed by Apps </v-list-item-title>
          </v-list-item-content>
        </template>
        <v-row>
          <v-col
            v-for="ts in appInserters"
            :key="ts.id"
            cols="12"
            sm="12"
            md="6"
            lg="6"
            xl="4"
          >
            <insert-card :object="ts" />
          </v-col>
        </v-row>
      </v-list-group>
    </v-list>
  </h-page-container>
</template>
<script>
function hasWrite(o) {
  let access = o.access.split(" ");
  return (
    o.meta.schema.type !== undefined &&
    (access.includes("*") || access.includes("write"))
  );
}

import InsertCard from "./insert_card.vue";

export default {
  components: {
    InsertCard,
  },
  computed: {
    user() {
      return this.$store.state.app.info.user;
    },
    inserters() {
      return Object.keys(
        this.$store.state.heedy.userObjects[this.user.username] || {}
      )
        .map((id) => this.$store.state.heedy.objects[id])
        .filter((o) => o.type === "timeseries" && hasWrite(o));
    },
    yourInserters() {
      return this.inserters.filter((o) => o.app == null);
    },
    appInserters() {
      return this.inserters.filter((o) => o.app != null);
    },
  },
  created() {
    this.$store.dispatch("readUserObjects", { username: this.user.username });
  },
};
</script>