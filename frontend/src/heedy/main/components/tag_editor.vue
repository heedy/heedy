<template>
  <v-combobox
    outlined
    :delimiters="[' ']"
    :deletable-chips="true"
    label="Tags"
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
    value: String
  },
  computed: {
    items() {
      if (this.value == "") {
        return [];
      }
      return this.value.split(" ");
    }
  },
  methods: {
    oninput(event) {
      this.$emit("input", event.join(" "));
    }
  }
};
</script>