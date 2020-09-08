<template>
  <div style="width:100%" ref="widthdiv">
    <virtual-table
      v-if="tabledata.length==1"
      :minWidth="width"
      :height="600"
      :config="config[0].columns"
      :data="tabledata[0]"
    ></virtual-table>
    <v-tabs v-else v-model="tab">
      <v-tab v-for="(tval,i) in tabledata" :key="i">Series {{ i+1 }}</v-tab>
      <v-tab-item v-for="(tval,i) in tabledata" :key="i" :value="i">
        <virtual-table :minWidth="width" :height="600" :config="config[i].columns" :data="tval"></virtual-table>
      </v-tab-item>
    </v-tabs>
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
    tab: 0,
  }),
  computed: {
    tabledata() {
      // Generate objects for all the data
      let ndp = this.dataset.map((series) =>
        series.map((dp) => {
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
        })
      );
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
