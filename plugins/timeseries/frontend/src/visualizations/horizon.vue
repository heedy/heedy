<template>
  <div ref="chart" style="width: 100%"></div>
</template>
<script>
import { HorizonTSChart } from "../../dist/d3v.mjs";
import moment from "../../dist/moment.mjs";
export default {
  props: {
    data: Object,
  },
  methods: {
    handleResize(event) {
      this._tc.width(this.$refs.chart.clientWidth);
    },
  },
  beforeDestroy() {
    window.removeEventListener("resize", this.handleResize);
  },
  watch: {
    data(nd, old) {
      this._tc
        .data(nd.data)
        .horizonBands(nd.bands)
        .seriesLabelFormatter(nd.label ? (x) => x : () => "");
    },
  },
  mounted() {
    window.addEventListener("resize", this.handleResize);
    console.vlog("MOUNT", this.data);
    this._tc = HorizonTSChart()(this.$refs.chart)
      .interpolationCurve(false)
      .enableZoom(true)
      .series("series")
      .height(600)
      .horizonBands(this.data.bands)
      .yNormalize(true)
      .seriesLabelFormatter(this.data.label ? (x) => x : () => "")
      .tooltipContent(
        ({ ts, val, series }) =>
          `<b>${series} ${moment(ts).format("LTS")}</b>: ${Math.round(val)}`
      );

    this._tc.width(this.$refs.chart.clientWidth).data(this.data.data);
  },
};
</script>
