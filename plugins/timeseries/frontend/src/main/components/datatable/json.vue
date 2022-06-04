<template>
  <v-tooltip bottom>
    <template v-slot:activator="{ on, attrs }">
      <div
        :style="{
          width:'100%',
          overflowX: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
          textAlign: align,
        }"
        v-bind="attrs"
        v-on="on"
      >
        <pre v-html="highlighted"></pre>
      </div>
    </template>
    <pre
      v-html="highlightedIndent"
      style="
        background: #f0f0f0;
        color: black;
        padding-left: 5px;
        padding-right: 10px;
      "
    ></pre>
  </v-tooltip>
</template>
<script>
import { hljs } from "../../../../dist/markdown-it.mjs";

export default {
  props: {
    value: null,
    column: Object,
    align: {
      type: String,
      default: "center",
    },
  },
  computed: {
    highlighted() {
      if (this.value === undefined) {
        return "";
      }
      const hlv = hljs.highlight(JSON.stringify(this.value), {
        language: "json",
      }).value;
      return hlv;
    },
    highlightedIndent() {
      if (this.value === undefined) {
        return "";
      }
      return hljs.highlightAuto(JSON.stringify(this.value, null, 2)).value;
    },
  },
};
</script>