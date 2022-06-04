<template>
  <v-tooltip bottom :disabled="!dotted">
    <template v-slot:activator="{ on, attrs }">
      <span v-bind="attrs" v-on="on">{{ text }}</span>
    </template>
    <span>{{ value }}</span>
  </v-tooltip>
</template>
<script>
export default {
  props: {
    value: Number,
    column: Object,
    align: {
      type: String,
      default: "center",
    },
  },
  data: () => ({
    dotted: false,
  }),
  methods: {
    addDots(news, converter = (x) => x) {
      const oldf = this.value;
      const newf = parseFloat(news);
      if (oldf == newf) {
        return newf;
      }
      this.dotted = true;
      return `${converter(newf)}...`;
    },
  },
  computed: {
    text: function () {
      if (this.value === undefined) {
        return "";
      }
      if (Number.isInteger(this.value)) {
        let sval = this.value.toLocaleString();
        if (sval.length <= 11) {
          return sval;
        }
        // Otherwise, return it in exponential notation
        return this.addDots(this.value.toExponential(4), (x) =>
          x.toExponential()
        );
      }

      // It a float
      if (Math.abs(this.value) < 1e5) {
        if (Math.abs(this.value) >= 1.0) {
          return this.addDots(this.value.toFixed(2));
        }
        if (Math.abs(this.value) >= 1e-4) {
          return this.addDots(this.value.toFixed(6));
        }
      }
      return this.addDots(this.value.toExponential(5), (x) =>
        x.toExponential()
      );
    },
  },
};
</script>