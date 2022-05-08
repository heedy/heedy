<template>
  <v-form @submit="update" v-model="formValid" style="padding: 20px">
    <h-jsf :schema="schema" v-model="modified" />
    <v-card-actions>
      <v-spacer></v-spacer>

      <v-btn dark small color="info" type="submit"
        >Update {{ plugin }} Settings</v-btn
      >
    </v-card-actions>
  </v-form>
</template>
<script>
export default {
  data: () => ({
    formValid: false,
    modified: {},
  }),
  props: {
    schema: Object,
    value: Object,
    plugin: String,
  },
  methods: {
    update: async function (event) {
      event.preventDefault();
      if (!this.formValid) {
        return;
      }

      let modified = { ...this.modified };
      // Need to check if any keys are being removed
      Object.keys(this.value).forEach((k) => {
        if (!modified.hasOwnProperty(k)) {
          modified[k] = null; // Explicitly delete the variable
        }
      });
      
      this.$emit("update", modified);
    },
  },
  watch: {
    value(newVal) {
      this.modified = JSON.parse(JSON.stringify(newVal));
    },
  },
  created() {
    this.modified = JSON.parse(JSON.stringify(this.value));
  },
};
</script>