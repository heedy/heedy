<template>
  <div style="width:100%" ref="widthdiv">
    <virtual-table
      v-if="config.length==1"
      :minWidth="width"
      :height="600"
      :config="config[0].columns"
      :data="config[0].data"
    ></virtual-table>
    <v-tabs v-else v-model="tab">
      <v-tab v-for="(tval,i) in config" :key="i">Series {{ i+1 }}</v-tab>
      <v-tab-item v-for="(tval,i) in config" :key="i" :value="i">
        <virtual-table
          :minWidth="width"
          :height="600"
          :config="config[i].columns"
          :data="config[i].data"
        ></virtual-table>
      </v-tab-item>
    </v-tabs>
  </div>
</template>
<script>
import VirtualTable from "vue-virtual-table";

export default {
  components: {
    VirtualTable,
  },
  props: {
    query: Array,
    config: Array,
  },
  data: () => ({
    width: 100,
    tab: 0,
  }),

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
