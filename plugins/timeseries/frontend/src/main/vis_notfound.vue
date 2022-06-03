<template>
  <v-alert
    text
    outlined
    color="deep-orange"
    icon="error_outline"
  >Visualization display of type "{{ type }}" not found</v-alert>
</template>
<script>
export default {
  props: {
    query: Object,
    config: {},
    data: {},
    type: String,
  }
}
</script>
