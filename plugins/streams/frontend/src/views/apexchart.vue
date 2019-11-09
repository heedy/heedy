<template>
  <apex-chart width="100%" :type="data.type" :options="data.chartOptions" :series="data.series"></apex-chart>
</template>
<script>
import ApexChart from "../../dist/apexcharts.mjs";
export default {
  components: {
    ApexChart
  },
  props: {
    data: Object
  }
};
</script>