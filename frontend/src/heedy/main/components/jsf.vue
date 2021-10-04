<template>
  <v-jsf
    class="markdownview"
    :schema="schema"
    :options="opts"
    :value="value"
    @input="(e) => $emit('input', e)"
  />
</template>
<script>
import { md } from "../../../dist/markdown-it.mjs";

export default {
  props: {
    value: {
      required: true,
    },
    schema: {
      type: Object,
      required: true,
    },
    options: {
      type: Object,
      default: () => ({}),
    },
  },
  computed: {
    opts() {
      return {
        markdown: (r) => {
          if (r === undefined || r == null || r == "") {
            return null;
          }
          return md.render(r);
        },
        ...this.options,
      };
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