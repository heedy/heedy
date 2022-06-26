<template>
  <span>{{ textTime }}</span>
</template>
<script>
export default {
  props: {
    value: Number | Date,
    column: Object,
    align: {
      type: String,
      default: "center",
    },
  },
  computed: {
    textTime: function () {
      if (typeof this.value === "number") {
        return new Date(this.value * 1000).toLocaleString();
      }
      return this.value.toLocaleString();
    },
  },
};
</script>