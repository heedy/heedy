<template>
  <span v-html="rendered" class="markdownview"></span>
</template>
<script>
import { md } from "../../../dist/markdown-it.mjs";

export default {
  props: {
    value: {
      type: String,
      default: "",
    },
  },
  computed: {
    rendered() {
      if (this.value == "") {
        return "";
      }
      return md.render(this.value);
    },
  },
};
</script>

<style>
.markdownview p {
  padding-top: 15px;
}
.markdownview h1 {
  padding-top: 15px;
}
.markdownview h2 {
  padding-top: 15px;
}
.markdownview h3 {
  padding-top: 15px;
}
.markdownview h4 {
  padding-top: 15px;
}
.markdownview img {
  max-width: 100%;
}
</style>