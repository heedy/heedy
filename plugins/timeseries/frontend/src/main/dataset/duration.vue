<template>
  <v-select
    prepend-icon="update"
    :items="items"
    :value="value"
    :search-input.sync="search"
    label="Each"
    @input="oninput"
  ></v-select>
</template>
<script>
export default {
  props: {
    value: Number,
  },
  data: () => ({
    items: [
      {
        text: "1 second",
        value: 1,
      },
      {
        text: "1 minute",
        value: 60,
      },
      {
        text: "5 minutes",
        value: 60 * 5,
      },
      {
        text: "1 hour",
        value: 60 * 60,
      },
      {
        text: "1 day",
        value: 60 * 60 * 24,
      },
      {
        text: "1 week",
        value: 60 * 60 * 24 * 7,
      },
    ],
    search: "",
  }),
  methods: {
    oninput(v) {
      this.$emit("input", v);
    },
  },
  watch: {
    value(v) {
      if (!this.items.map((i) => i.value).includes(v)) {
        this.items.push({
          text: `${v} seconds`,
          value: v,
        });
      }
    },
  },
  created() {
    if (!this.items.map((i) => i.value).includes(this.value)) {
      this.items.push({
        text: `${this.value} seconds`,
        value: this.value,
      });
    }
  },
};
</script>
