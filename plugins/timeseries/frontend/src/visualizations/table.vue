<template>
  <div style="width: 100%" ref="widthdiv">
    <h-table
      v-if="config.length == 1"
      :columns="config[0].columns"
      :data="config[0].data"
    />
    <v-tabs v-else v-model="tab">
      <v-tab v-for="(tval, i) in config" :key="i">{{ tval.label }}</v-tab>
      <v-tab-item v-for="(tval, i) in config" :key="i" :value="i">
        <h-table :columns="tval.columns" :data="config[i].data" />
      </v-tab-item>
    </v-tabs>
  </div>
</template>
<script>
export default {
  props: {
    query: Object,
    config: Array,
  },
  data: () => ({
    tab: 0,
  }),
};
</script>
