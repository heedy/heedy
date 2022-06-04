<template>
  <div style="width: 100%" ref="widthdiv">
    <h-table
      v-if="config.length == 1"
      :columns="data[0].columns"
      :data="data[0].data"
    />
    <v-tabs v-else v-model="tab">
      <v-tab v-for="(tval, i) in data" :key="i">{{ tval.label }}</v-tab>
      <v-tab-item v-for="(tval, i) in data" :key="i" :value="i">
        <h-table :columns="tval.columns" :data="tval.data" />
      </v-tab-item>
    </v-tabs>
  </div>
</template>
<script>
export default {
  props: {
    query: Object,
    config: Array,
    data: Array,
    type: String
  },
  data: () => ({
    tab: 0,
  }),
};
</script>
