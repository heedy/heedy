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
      <v-btn outlined @click="setBackup(false)">Restart</v-btn>
      <div class="flex-grow-1"></div>
      <v-btn color="primary" dark class="mb-2" @click="update">Save</v-btn>
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
    setBackup: async function (newValue) {
      let o = this.$store.state.heedy.updates.options;
      if (o == null) {
        o = {
          backup: false,
          deleted: [],
        };
      }
      o = {
        ...o,
        backup: o.backup?true:newValue,
      };
      let res = await this.$frontend.rest(
        "POST",
        "api/server/updates/options",
        o
      );
      if (!res.response.ok) {
        console.verror("Update error: ", res.data.error_description);
        this.alert = res.data.error_description;
      }
      this.$store.dispatch("getUpdates");
    },
    reload: async function() {
      try {
        let res = await fetch("/api/server/updates/heedy.conf", {
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
      let res = await fetch("/api/server/updates/heedy.conf", {
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