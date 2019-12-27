<template>
  <div style="width:100%" ref="widthdiv">
    <virtual-table
      :minWidth="width"
      :height="600"
      :config="data.config"
      :data="data.data"
    >
    </virtual-table>
  </div>
</template>
<script>
import VirtualTable from "vue-virtual-table";

export default {
  components: {
    VirtualTable
  },
  props: {
    data: Object
  },
  data: () => ({
    width: 100
  }),
  methods: {
    handleResize(event) {
      this.width = this.$refs.widthdiv.clientWidth - 2;
    }
  },
  beforeDestroy() {
    window.removeEventListener("resize", this.handleResize);
  },
  mounted() {
    window.addEventListener("resize", this.handleResize);
    this.width = this.$refs.widthdiv.clientWidth - 2;
  }
};
</script>
