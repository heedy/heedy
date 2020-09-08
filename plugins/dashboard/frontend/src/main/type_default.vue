<template>
  <v-alert text outlined color="deep-orange" icon="error_outline"
    >Unrecognized element type "{{ element.type }}"</v-alert
  >
</template>
<script>
export default {
  props: {
    object: Object,
    element: Object,
  },
};
</script>
