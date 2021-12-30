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
          <span
            v-if="
              item[col.prop + '.type'] === undefined ||
              item[col.prop + '.type'] == ''
            "
            >{{ item[col.prop] }}</span
          >
          <component
            v-else
            :is="getColumn(item[col.prop + '.type']).component"
            :value="item[col.prop]"
            :column="col"
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
    getColumn(t) {
      return getCol(t);
    },
  },
};
</script>