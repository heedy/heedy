<template>
  <v-simple-table fixed-header :height="height">
    <thead>
      <tr>
        <th v-for="column in columns" :key="column.prop">
          {{ column.name }}
        </th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="(item, i) in data" :key="i">
        <td v-for="col in columns" :key="col.prop">
          <span v-if="item[col.prop + '.type']===undefined && typeof item[col.prop]=='string'">{{ item[col.prop] }}</span>
          <component v-else
            :is="getColumn(item[col.prop],item[col.prop + '.type']).component"
            :value="item[col.prop]"
            :column="col"
            align="left"
          />
        </td>
      </tr>
    </tbody>
  </v-simple-table>
</template>
<script>
import getCol from "./datatable/columns.js";
export default {
  props: {
    data: {
      type: Array,
      required: true,
    },
    columns: {
      type: Array,
      required: true,
    },
  },
  computed: {
    height() {
      if (this.data.length >= 10) {
        return 9 * 48 + 48 + 24;
      }
      return 48 * this.data.length + 48;
    },
  },
  methods: {
    getColumn(datapoint,t) {
      if (t!==undefined) {
        return getCol(t);
      }
      return getCol(typeof datapoint);
    },
  },
};
</script>