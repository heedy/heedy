<template>
  <v-select prepend-icon="show_chart" :items="items" :value="value" @input="(v)=> $emit('input',v)"></v-select>
</template>
<script>
export default {
  props: {
    value: String
  },
  data: () => ({
    items: ["closest", "sum", "count"]
  })
};
</script>