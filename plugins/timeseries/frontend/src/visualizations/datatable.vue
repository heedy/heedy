<template>
  <div style="width: 100%">
    <h-timeseries-datatable
      v-if="config.length == 1"
      :data="config[0].data"
      :columns="config[0].columns"
      :timeseries="config[0].timeseries"
      :editable="config[0].editable"
    />
    <v-tabs v-else v-model="tab">
      <v-tab v-for="(tval, i) in config" :key="i">{{ config[i].label }}</v-tab>
      <v-tab-item v-for="(tval, i) in config" :key="i" :value="i">
        <h-timeseries-datatable
          :columns="config[i].columns"
          :data="config[i].data"
          :editable="config[i].editable"
          :timeseries="config[i].timeseries"
        />
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