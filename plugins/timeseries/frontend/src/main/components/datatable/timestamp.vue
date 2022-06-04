<template>
  <span>{{ textTime }}</span>
</template>
<script>
export default {
  props: {
    value: Number,
    column: Object,
    align: {
      type: String,
      default: "center"
    }
  },
  computed: {
    textTime: function () {
      return new Date(this.value * 1000).toLocaleString();
    },
  },
};
</script>