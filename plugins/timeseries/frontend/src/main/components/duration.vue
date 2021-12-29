<template>
  <v-text-field :label="label" v-model="text" :error="!canParse" />
</template>
<script>
const minute = 60;
const hour = 60 * minute;
const day = 24 * hour;
const week = 7 * day;
const month = 30 * day;
const year = 365 * day;
const times = [
  [year, "y"],
  [month, "mo"],
  [week, "w"],
  [day, "d"],
  [hour, "h"],
  [minute, "m"],
  [1, "s"],
];

function num2text(num) {
  if (isNaN(num) || num < 0 || !isFinite(num)) {
    return "";
  }
  if (num === 0) {
    return "0s";
  }
  let res = "";
  if (num < 0) {
    res += "-";
    num = -num;
  }
  let tval = num;
  for (let i = 0; i < times.length - 1; i++) {
    const [val, unit] = times[i];
    const n = Math.floor(tval / val);
    if (n > 0) {
      res += `${n}${unit}`;
      tval -= n * val;
    }
  }
  if (tval != 0) {
    res += `${tval}s`;
  }
  return res;
}

function text2num(txt) {
  txt = txt.toLowerCase().trim();
  let res = 0;
  let sign = 1;
  if (txt.startsWith("-")) {
    sign = -1;
    txt = txt.substr(1);
  }
  for (let i = 0; i < times.length; i++) {
    const [val, unit] = times[i];
    const unitIndex = txt.indexOf(unit);
    if (unitIndex > -1) {
      const str = txt.substr(0, unitIndex);
      const n = parseFloat(str);
      if (isNaN(str) || n < 0 || !isFinite(n)) {
        return NaN;
      }
      res += n * val;
      txt = txt.substr(unitIndex + unit.length);
    }
  }
  // Add any remaining as seconds
  if (txt.length > 0) {
    const n = parseFloat(txt);
    if (isNaN(txt) || n < 0 || !isFinite(n)) {
      return NaN;
    }
    res += n;
  }
  return sign * res;
}

export default {
  props: {
    value: Number,
    column: Object,
    label: {
      type: String,
      default: "Duration",
    },
    allowNegative: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      text: num2text(this.value),
      canParse: true,
      currentValue: this.value,
    };
  },
  watch: {
    text(val) {
      let num = text2num(val);
      this.canParse = !isNaN(num) && (this.allowNegative || num >= 0);
      if (!this.canParse) {
        num = NaN;
      }
      this.currentValue = num;
      this.$emit("input", num);
    },
    value(v) {
      // Only update the text if the value is different - the user can format it how they want
      if (
        this.currentValue != v &&
        (!isNaN(v) || isNaN(v) != isNaN(this.currentValue))
      ) {
        this.currentValue = this.value;
        this.text = num2text(this.value);
        this.canParse =
          !isNaN(this.value) && (this.allowNegative || this.value >= 0);
      }
    },
  },
};
</script>