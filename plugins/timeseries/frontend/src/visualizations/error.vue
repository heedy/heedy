<template>
  <v-alert text outlined color="deep-orange" icon="error_outline">
    <div class="text-h6">{{ data.title }}</div>
    <p>{{ data.error }}</p>
    <codemirror
        v-if="data.showConfig!==undefined&&data.showConfig"
            :value="JSON.stringify(config, null, '  ')"
            :options="cmOptions"
          ></codemirror>
  </v-alert>
</template>
<script>
export default {
  props: {
    query: Object,
    config: {},
    data: Object,
    type: String,
  },
  data: () => ({
      cmOptions: {
      tabSize: 2,
      mode: "text/javascript",
      readOnly: true,
    },
  })
};
</script>
<style>
.CodeMirror {
  border: 1px solid #eee;
  height: auto;
}
.CodeMirror-scroll {
    max-height: 200px;
}
</style>