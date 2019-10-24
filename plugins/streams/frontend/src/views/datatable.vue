<template>
  <div ref="tablediv" style="max-height: 400px; overflow-y: auto">
    <table style="width: 100%; overflow: auto">
      <thead class="v-data-table-header">
        <tr>
          <th class="text-start">Timestamp</th>
          <th class="text-start" v-for="h in data.header" :key="h">{{ h }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="d in data.data" :key="d.key">
          <td class="text-start">{{ ts(d.t) }}</td>
          <td class="text-start" v-for="dd of d.d.entries()" :key="dd[0]">{{ dd[1] }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
<script>
import moment from "../../dist/moment.mjs";
export default {
  props: {
    data: Object
  },
  methods: {
    ts(t) {
      return moment.unix(t).calendar();
    }
  },
  mounted() {
    let td = this.$refs.tablediv;
    td.scrollTop = td.scrollHeight;
  }
};
</script>
<style scoped>
tr:nth-child(even) {
  background-color: #f2f2f2;
}
</style>