<template>
  <div ref="chart" style="width: 100%"></div>
</template>
<script>
import { TimelinesChart } from "../../dist/d3v.mjs";
export default {
  props: {
    query: Object,
    config: Object,
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
        .zQualitative(true)
        .leftMargin(nd.leftMargin)
        .rightMargin(nd.rightMargin)
        .timeFormat(nd.timeFormat)
        .data(nd.data);
    },
  },
  mounted() {
    window.addEventListener("resize", this.handleResize);
    this._tc = TimelinesChart()
      .enableOverview(false)
      .enableAnimations(false)
      .topMargin(35)
      .maxLineHeight(40)(this.$refs.chart);

    this._tc
      .width(this.$refs.chart.clientWidth)
      .zQualitative(true)
      .leftMargin(this.config.leftMargin)
      .rightMargin(this.config.rightMargin)
      .timeFormat(this.config.timeFormat)
      .data(this.config.data);
  },
};
</script>
