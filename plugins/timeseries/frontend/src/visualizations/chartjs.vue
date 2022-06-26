<template>
  <div>
    <div v-for="(d, idx) in data" :key="idx">
      <linechart
        v-if="d.type == 'line'"
        :chartData="d.data"
        :chartOptions="d.options"
        :width="null"
        :height="null"
      />
      <barchart
        v-else-if="d.type == 'bar'"
        :chartData="d.data"
        :chartOptions="d.options"
        :width="null"
        :height="null"
      />
      <horizontalbarchart
        v-else-if="d.type == 'horizontalBar'"
        :chartData="d.data"
        :chartOptions="d.options"
        :width="null"
        :height="null"
      />
      <doughnutchart
        v-else-if="d.type == 'doughnut'"
        :chartData="d.data"
        :chartOptions="d.options"
        :width="null"
        :height="null"
      />
      <piechart
        v-else-if="d.type == 'pie'"
        :chartData="d.data"
        :chartOptions="d.options"
        :width="null"
        :height="null"
      />
      <polarareachart
        v-else-if="d.type == 'polarArea'"
        :chartData="d.data"
        :chartOptions="d.options"
        :width="null"
        :height="null"
      />
      <radarchart
        v-else-if="d.type == 'radar'"
        :chartData="d.data"
        :chartOptions="d.options"
        :width="null"
        :height="null"
      />
      <bubblechart
        v-else-if="d.type == 'bubble'"
        :chartData="d.data"
        :chartOptions="d.options"
        :width="null"
        :height="null"
      />
      <scatterchart
        v-else-if="d.type == 'scatter'"
        :chartData="d.data"
        :chartOptions="d.options"
        :width="null"
        :height="null"
      />
      <div v-else>Unrecognized chart type</div>
    </div>
  </div>
</template>
<script>
import Chartjs from "../../dist/chartjs.mjs";

/*
let getChart = (c) => ({
  extends: c,
});
*/
const getChart = (c) => c;

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
    query: Object,
    config: {},
    data: {},
    type: String
  },
};
</script>
