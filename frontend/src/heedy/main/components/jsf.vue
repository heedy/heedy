<template>
  <v-jsf
    class="markdownview"
    :schema="schema"
    :options="opts"
    :value="value"
    @input="(e) => $emit('input', e)"
  >
    <template
      v-for="ins in schemaFormElements"
      :slot="`custom-` + ins.k"
      slot-scope="{ value, label, on }"
    >
      <component
        :key="ins.k"
        :is="ins.v"
        :value="value"
        v-on="on"
        :label="label"
      />
    </template>
  </v-jsf>
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
    schemaFormElements() {
      const el = this.$store.state.heedy.schema_form_elements;
      return Object.keys(el).map((k) => ({ k: k, v: el[k] }));
    },
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