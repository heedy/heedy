<template>
  <div>
    <div v-for="(data, idx) in config.charts" :key="idx">
      <linechart
        v-if="data.type == 'line'"
        :chartData="data.data"
        :options="data.options"
        :width="null"
        :height="null"
      />
      <barchart
        v-else-if="data.type == 'bar'"
        :chartData="data.data"
        :options="data.options"
        :width="null"
        :height="null"
      />
      <horizontalbarchart
        v-else-if="data.type == 'horizontalBar'"
        :chartData="data.data"
        :options="data.options"
        :width="null"
        :height="null"
      />
      <doughnutchart
        v-else-if="data.type == 'doughnut'"
        :chartData="data.data"
        :options="data.options"
        :width="null"
        :height="null"
      />
      <piechart
        v-else-if="data.type == 'pie'"
        :chartData="data.data"
        :options="data.options"
        :width="null"
        :height="null"
      />
      <polarareachart
        v-else-if="data.type == 'polarArea'"
        :chartData="data.data"
        :options="data.options"
        :width="null"
        :height="null"
      />
      <radarchart
        v-else-if="data.type == 'radar'"
        :chartData="data.data"
        :options="data.options"
        :width="null"
        :height="null"
      />
      <bubblechart
        v-else-if="data.type == 'bubble'"
        :chartData="data.data"
        :options="data.options"
        :width="null"
        :height="null"
      />
      <scatterchart
        v-else-if="data.type == 'scatter'"
        :chartData="data.data"
        :options="data.options"
        :width="null"
        :height="null"
      />
      <div v-else>Unrecognized chart type</div>
    </div>
  </div>
</template>
<script>
import Chartjs from "../../dist/chartjs.mjs";

let getChart = (c) => ({
  extends: c,
  mixins: [Chartjs.mixins.reactiveProp],
  props: {
    chartData: Object,
    options: Object,
  },
  watch: {
    options: {
      handler(newOption, oldOption) {
        this.$data._chart.destroy();
        this.renderChart(this.chartData, this.options);
      },
      deep: true,
    },
  },
  mounted() {
    this.renderChart(this.chartData, this.options);
  },
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
    scatterchart: getChart(Chartjs.Scatter),
  },
  props: {
    query: Array,
    config: Object,
  },
};
</script>
