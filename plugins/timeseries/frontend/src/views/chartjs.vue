<template>
  <linechart v-if="data.type=='line'" :chartData="data.data" :options="data.options" />
  <barchart v-else-if="data.type=='bar'" :chartData="data.data" :options="data.options" />
  <horizontalbarchart
    v-else-if="data.type=='horizontalBar'"
    :chartData="data.data"
    :options="data.options"
  />
  <doughnutchart v-else-if="data.type=='doughnut'" :chartData="data.data" :options="data.options" />
  <piechart v-else-if="data.type=='pie'" :chartData="data.data" :options="data.options" />
  <polarareachart
    v-else-if="data.type=='polarArea'"
    :chartData="data.data"
    :options="data.options"
  />
  <radarchart v-else-if="data.type=='radar'" :chartData="data.data" :options="data.options" />
  <bubblechart v-else-if="data.type=='bubble'" :chartData="data.data" :options="data.options" />
  <scatterchart v-else-if="data.type=='scatter'" :chartData="data.data" :options="data.options" />
  <div v-else>Unrecgnized chart type</div>
</template>
<script>
import Chartjs from "../../dist/chartjs.mjs";

let getChart = c => ({
  extends: c,
  mixins: [Chartjs.mixins.reactiveProp],
  props: {
    chartData: Object,
    options: Object
  },
  watch: {
    options: {
      handler(newOption, oldOption) {
        this.$data._chart.destroy();
        this.renderChart(this.chartData, this.options);
      },
      deep: true
    }
  },
  mounted() {
    console.log(this.chartData, this.options);
    this.renderChart(this.chartData, this.options);
  }
});

export default {
  components: {
    linechart: getChart(Chartjs.Line),
    barchart: getChart(Chartjs.Bar),
    horizontalbarchart: getChart(Chartjs.HorizontalBar),
    doughnutchart: getChart(Chartjs.Doughnut),
    piechart: getChart(Chartjs.Pie),
    polarareachart: getChart(Chartjs.PolarArea),
    radarchart: getChart(Chartjs.Radar),
    bubblechart: getChart(Chartjs.Bubble),
    scatterchart: getChart(Chartjs.Scatter)
  },
  props: {
    data: Object
  }
};
</script>