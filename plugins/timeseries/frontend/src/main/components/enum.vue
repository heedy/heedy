<template>
  <v-combobox
    outlined
    :delimiters="[' ']"
    :deletable-chips="true"
    label="Event Types"
    multiple
    small-chips
    :value="items"
    @input="oninput"
    :hide-no-data="true"
  ></v-combobox>
</template>
<script>
export default {
  props: {
    value: Object,
  },
  computed: {
    items() {
      return this.value.enum;
    },
  },
  methods: {
    oninput(event) {
      this.$emit("input", { ...this.value, enum: event });
    },
  },
};
</script>