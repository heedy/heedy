<template>
  <span>{{ textTime }}</span>
</template>
<script>
export default {
  props: {
    value: Number,
    column: Object,
  },
  computed: {
    textTime: function () {
      return new Date(this.value * 1000).toLocaleString();
    },
  },
};
</script>