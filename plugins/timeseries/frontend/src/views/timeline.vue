<template>
  <div ref="chart" style="width: 100%"></div>
</template>
<script>
import TimelinesChart from "../../dist/timelines-chart.mjs";
export default {
  props: {
    data: Object
  },
  methods: {
    handleResize(event) {
      console.log("REESIZE");
      this._tc.width(this.$refs.chart.clientWidth);
    }
  },
  beforeDestroy() {
    window.removeEventListener("resize", this.handleResize);
  },
  watch: {
    data(nd, old) {
      this._tc
        .zQualitative(nd.discrete)
        .leftMargin(nd.leftMargin)
        .rightMargin(nd.rightMargin)
        .timeFormat(nd.timeFormat)
        .data(nd.data);
    }
  },
  mounted() {
    window.addEventListener("resize", this.handleResize);
    console.log("MOUNT", this.data);
    this._tc = TimelinesChart()
      .enableOverview(false)
      .enableAnimations(false)
      .topMargin(35)
      .maxLineHeight(40)(this.$refs.chart);

    this._tc
      .width(this.$refs.chart.clientWidth)
      .zQualitative(this.data.discrete)
      .leftMargin(this.data.leftMargin)
      .rightMargin(this.data.rightMargin)
      .timeFormat(this.data.timeFormat)
      .data(this.data.data);
  }
};
</script>
