<template>
  <v-form @submit="update" v-model="formValid" style="padding: 20px">
    <h-jsf :schema="schema" v-model="modified" />
    <v-card-actions>
      <v-spacer></v-spacer>

      <v-btn dark small color="info" type="submit" :loading="loading"
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
    loading: false,
  }),
  props: {
    schema: Object,
    value: Object,
    plugin: String,
  },
  methods: {
    update: async function (event) {
      event.preventDefault();
      if (this.loading) return;
      if (!this.formValid) {
        return;
      }
      this.loading = true;

      console.vlog(`Updating settings for ${this.plugin}`, this.modified);

      let modified = { ...this.modified };
      // Need to check if any keys are being removed
      Object.keys(this.value).forEach((k) => {
        if (!modified.hasOwnProperty(k)) {
          modified[k] = null; // Explicitly delete the variable
        }
      });
      if (Object.keys(modified).length > 0) {
        let result = await this.$frontend.rest(
          "PATCH",
          `api/users/${encodeURIComponent(
            this.$store.state.app.info.user.username
          )}/settings/${encodeURIComponent(this.plugin)}`,
          modified
        );

        if (!result.response.ok) {
          this.$store.commit("alert", {
            type: "error",
            text: result.data.error_description,
          });
          this.loading = false;
          return;
        }
      }
      this.loading = false;
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