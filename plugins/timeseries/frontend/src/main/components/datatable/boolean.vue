<template>
  <span>{{ value }}</span>
</template>
<script>
export default {
  props: {
    value: Boolean,
    column: Object,
  },
};
</script>