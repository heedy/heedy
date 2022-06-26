<template>
  <div>
    <v-tabs v-model="showTab" grow>
      <v-tab> Config </v-tab>
      <v-tab> Data </v-tab>
      <v-tab> Query </v-tab>
    </v-tabs>
    <v-tabs-items v-model="showTab" grow>
      <v-tab-item>
        <codemirror
          :value="JSON.stringify(config, null, '  ')"
          :options="cmOptions"
        ></codemirror>
      </v-tab-item>
      <v-tab-item>
        <codemirror
          :value="JSON.stringify(data, null, '  ')"
          :options="cmOptions"
        ></codemirror>
      </v-tab-item>
      <v-tab-item>
        <codemirror
          :value="JSON.stringify(query, null, '  ')"
          :options="cmOptions"
        ></codemirror>
      </v-tab-item>
    </v-tabs-items>
  </div>
</template>
<script>
export default {
  props: {
    query: Object,
    config: {},
    data: {},
    type: String,
  },
  data: () => ({
    showTab: 0,
    cmOptions: {
      tabSize: 2,
      mode: "text/javascript",
      readOnly: true,
    },
  }),
};
</script>
<style>
.CodeMirror {
  border: 1px solid #eee;
  height: auto;
}
.CodeMirror-scroll {
  max-height: 600px;
}
.v-alert__content {
  overflow: hidden;
}
</style>