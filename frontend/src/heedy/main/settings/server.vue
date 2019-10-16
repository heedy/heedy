<template>
  <div>
    <v-flex v-if="alert.length>0">
      <div style="padding: 10px; padding-bottom: 0;">
        <v-alert text outlined color="deep-orange" icon="error_outline">{{ alert }}</v-alert>
      </div>
    </v-flex>
    <v-card-text>
      <codemirror :options="cmOptions" v-model="config" />
    </v-card-text>
    <v-card-actions>
      <div class="flex-grow-1"></div>
      <v-btn color="primary" dark class="mb-2" @click="update">Update</v-btn>
    </v-card-actions>
  </div>
</template>
<script>
export default {
  data: () => ({
    cmOptions: {
      autofocus: true,
      lineNumbers: true,
      extraKeys: {}
    },
    config: "Loading...",
    alert: ""
  }),
  methods: {
    reload: async function() {
      try {
        let res = await fetch("/api/heedy/v1/server/updates/heedy.conf", {
          method: "GET",
          credentials: "include",
          redirect: "follow"
        });

        this.config = await res.text();
      } catch (err) {
        this.config = "Failed to load configuration";
      }
    },
    update: async function() {
      let res = await fetch("/api/heedy/v1/server/updates/heedy.conf", {
        method: "POST",
        credentials: "include",
        redirect: "follow",
        body: this.config
      });
      if (!res.ok) {
        this.alert = (await res.json()).error_description;
        return;
      }
      this.alert = "";
      this.$store.dispatch("getUpdates");
    }
  },
  created() {
    this.reload();
    this.cmOptions.extraKeys["Ctrl-S"] = () => this.update();
    this.cmOptions.extraKeys["Cmd-S"] = () => this.update();
  }
};
</script>
<style>
.CodeMirror {
  height: auto;
  border: 1px solid #eee;
}
</style>