<template>
  <div style="width:100%" ref="widthdiv">
    <virtual-table
      :minWidth="width"
      :height="600"
      :config="config[0].columns"
      :data="data"
    >
    </virtual-table>
  </div>
</template>
<script>
import VirtualTable from "vue-virtual-table";
import moment from "../../dist/moment.mjs";

export default {
  components: {
    VirtualTable,
  },
  props: {
    query: Array,
    dataset: Array,
    config: Array,
  },
  data: () => ({
    width: 100,
  }),
  computed: {
    data() {
      // Generate objects for all the data
      let ndp = this.dataset[0].map((dp) => {
        let obj = {
          t: new Date(dp.t * 1000).toLocaleString(),
          d: JSON.stringify(dp.d),
          dt:
            dp.dt === undefined
              ? ""
              : moment.duration(dp.dt, "seconds").humanize(),
        };
        if (typeof dp.d == "object") {
          Object.keys(dp.d).map((k) => {
            obj[k] =
              typeof dp.d[k] !== "string" ? JSON.stringify(dp.d[k]) : dp.d[k];
          });
        }
        return obj;
      });
      console.log(ndp);
      return ndp;
    },
  },
  methods: {
    handleResize(event) {
      this.width = this.$refs.widthdiv.clientWidth - 2;
    },
  },
  beforeDestroy() {
    window.removeEventListener("resize", this.handleResize);
  },
  mounted() {
    window.addEventListener("resize", this.handleResize);
    this.width = this.$refs.widthdiv.clientWidth - 2;
  },
};
</script>
