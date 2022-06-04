<template>
  <span>{{ text }}</span>
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
];
export default {
  props: {
    value: Number,
    column: Object,
    align: {
      type: String,
      default: "center",
    },
  },
  computed: {
    text() {
      if (this.value === undefined) {
        return "";
      }
      let res = "";
      let tval = this.value;
      let units = 0;
      for (let i = 0; i < times.length; i++) {
        const [val, unit] = times[i];
        const n = Math.floor(tval / val);
        if (n > 0) {
          res += `${n}${unit}`;
          tval -= n * val;
          units++;
          if (units >= 2) {
            return res;
          }
        }
      }
      if (tval > 0) {
        if (Number.isInteger(tval)) {
          res += `${tval}s`;
        } else {
          res += `${tval.toFixed(2)}s`;
        }
      }
      return res;
    },
  },
};
</script>