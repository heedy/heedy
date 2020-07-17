<template>
  <v-text-field
    prepend-icon="code"
    label="Transform"
    :value="value"
    @input="(v)=> $emit('input',v)"
    :hide-details="true"
  ></v-text-field>
</template>
<script>
export default {
  props: {
    value: String
  }
};
</script>