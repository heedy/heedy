<template>
  <v-text-field
    prepend-icon="code"
    label="Transform"
    :value="value"
    @input="onInput"
    :hide-details="true"
    style="padding-bottom: 20px; font-family: monospace"
  ></v-text-field>
</template>
<script>
export default {
  props: {
    value: String,
  },
  methods: {
    onInput(v) {
      // Fix unicode “”„ that are auto-filled on macs/ipads
      v = v.replace(/[“”„]/gu, '"');

      // Same for ‘’
      v = v.replace(/[‘’]/gu, "'");
      this.$emit("input", v);
    },
  },
};
</script>