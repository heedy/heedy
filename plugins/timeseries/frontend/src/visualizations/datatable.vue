<template>
  <div style="width: 100%">
    <h-timeseries-datatable
      v-if="data.length == 1"
      :data="data[0].data"
      :columns="data[0].columns"
      :timeseries="data[0].timeseries"
      :editable="data[0].editable"
    />
    <v-tabs v-else v-model="tab">
      <v-tab v-for="(tval, i) in data" :key="i">{{ data[i].label }}</v-tab>
      <v-tab-item v-for="(tval, i) in data" :key="i" :value="i">
        <h-timeseries-datatable
          :columns="data[i].columns"
          :data="data[i].data"
          :editable="data[i].editable"
          :timeseries="data[i].timeseries"
        />
      </v-tab-item>
    </v-tabs>
  </div>
</template>
<script>
export default {
  props: {
    query: Object,
    config: Object,
    data: Array
  },
  data: () => ({
    tab: 0,
  }),
};
</script>