<template>
  <v-card flat>
    <h5 :style="{ color: valid ? 'black' : 'red' }">JSON Schema</h5>
    <codemirror v-model="content" :options="cmOptions"></codemirror>
  </v-card>
</template>
<script>
import { deepEqual } from "../../../util.mjs";
export default {
  props: {
    value: Object,
  },
  data: () => ({
    code: "",
    valid: true,
    cmOptions: {
      tabSize: 2,
      smartIndent: true,
      mode: "text/javascript",
    },
  }),
  computed: {
    content: {
      get() {
        return this.code;
      },
      set(v) {
        this.code = v;
        try {
          let val = JSON.parse(v);
          this.valid = true;
          this.$emit("input", val);
        } catch (e) {
          this.valid = false;
          this.$emit("input", null);
        }
      },
    },
  },
  watch: {
    value(v) {
      if (v != null) {
        if (!deepEqual(JSON.parse(this.code), v)) {
          this.code = JSON.stringify(v, null, "  ");
        }
      }
    },
  },
  created() {
    this.code = JSON.stringify(this.value, null, "  ");
  },
};
</script>
<style>
.CodeMirror {
  border: 1px solid #eee;
  height: auto;
}
</style>