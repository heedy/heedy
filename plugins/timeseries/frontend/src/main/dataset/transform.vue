<template>
  <v-text-field
    label="Transform"
    :value="value"
    @input="onInput"
    :hide-details="true"
    style="padding-bottom: 20px; font-family: monospace"
  >
    <template v-slot:prepend>
      <v-tooltip bottom>
        <template v-slot:activator="{ on, attrs }">
          <v-btn
            icon
            style="padding: 0; margin: 0; margin-top: -5px; padding-top: 0"
            href="https://heedy.org/analysis/pipescript/"
            target="_blank"
            v-bind="attrs"
            v-on="on"
          >
            <v-icon>code</v-icon>
          </v-btn>
        </template>
        <span>PipeScript Transform Documentation</span>
      </v-tooltip>
    </template>
  </v-text-field>
</template>
<script>
export default {
  props: {
    value: String,
  },
  methods: {
    onInput(v) {
      // Fix unicode “”„ that are auto-filled on macs/ipads
      v = v.replace(/[“”„]/gu, '"');

      // Same for ‘’
      v = v.replace(/[‘’]/gu, "'");
      this.$emit("input", v);
    },
  },
};
</script>