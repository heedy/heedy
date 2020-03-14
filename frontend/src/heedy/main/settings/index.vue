<template >
  <h-card-page title="Settings">
    <v-tabs v-model="tab" show-arrows>
      <v-tabs-slider color="teal lighten-3"></v-tabs-slider>
      <v-tab
        v-for="r in routes"
        :key="r.path"
        :to="`/settings/${r.path}`"
      >{{ r.title!==undefined?r.title:r.path }}</v-tab>
    </v-tabs>
    <router-view></router-view>
    <v-flex v-if="hasUpdate>0">
      <div style="padding: 10px; padding-bottom: 0;">
        <v-alert outlined type="info" prominent border="left">
          <v-row align="center">
            <v-col class="grow">Heedy needs to restart to apply changes.</v-col>
            <v-col class="shrink" style="min-width: 190px;max-width: 100%;">
              <v-checkbox
                :input-value="hasBackup"
                value
                @change="setBackup"
                dense
                label="Backup Database"
                color="info"
              ></v-checkbox>
            </v-col>
            <v-col class="shrink">
              <v-btn color="info" style="width:100%" outlined @click="restart">Apply</v-btn>
              <v-btn color="info" style="width:100%" outlined @click="undoUpdates">Undo</v-btn>
            </v-col>
          </v-row>
        </v-alert>
      </div>
      <v-dialog v-model="restarting" persistent max-width="500px">
        <v-card>
          <v-card-title>Restarting...</v-card-title>
          <v-card-text>
            Please be patient while the server restarts. This may take several minutes if you enabled a plugin and/or have enabled database backup,
            since heedy might need to install plugin dependencies or copy all of your data to backup. This page will automatically reload once
            the server has restarted.
          </v-card-text>
        </v-card>
      </v-dialog>
    </v-flex>
    <v-flex v-if="alert.length >0">
      <div style="padding: 10px; padding-bottom: 0;">
        <v-alert outlined type="error" prominent border="left">
          <v-row align="center">
            <v-col class="grow">
              <b>Update Failed:</b>
              {{ alert }}
            </v-col>
            <v-col class="shrink">
              <v-btn color="error" outlined @click="alert=''">OK</v-btn>
            </v-col>
          </v-row>
        </v-alert>
      </div>
    </v-flex>
  </h-card-page>
</template>

<script>
export default {
  data: () => ({
    tab: null,
    restarting: false,
    alert: ""
  }),
  computed: {
    routes() {
      return this.$store.state.heedy.settings_routes;
    },
    hasUpdate() {
      let u = this.$store.state.heedy.updates;
      return u.heedy || u.config || u.plugins.length > 0 || u.options != null;
    },
    hasBackup() {
      let o = this.$store.state.heedy.updates.options;
      if (o == null) return false;
      return o.backup;
    }
  },
  methods: {
    setBackup: async function(newValue) {
      let o = this.$store.state.heedy.updates.options;
      if (o == null) {
        o = {
          backup: true,
          deleted: []
        };
      }
      o = {
        ...o,
        backup: newValue
      };
      let res = await this.$frontend.rest(
        "POST",
        "api/server/updates/options",
        o
      );
      if (!res.response.ok) {
        console.log("Update error: ", res.data.error_description);
        this.alert = res.data.error_description;
      }
      this.$store.dispatch("getUpdates");
    },
    undoUpdates: async function() {
      let res = await this.$frontend.rest("DELETE", "api/server/updates");
      if (!res.response.ok) {
        console.log("Update error: ", res.data.error_description);
        this.alert = res.data.error_description;
      } else {
        // Perform a refresh - undoing the update might have changed stuff in settings
        location.reload(true);
      }
    },
    restart: async function() {
      let res = await this.$frontend.rest("GET", "api/server/restart");

      this.restarting = true;

      function sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
      }
      await sleep(1000);

      res = await this.$frontend.rest("GET", "api/server/version");
      while (!res.response.ok) {
        await sleep(1000);
        res = await this.$frontend.rest("GET", "api/server/version");
      }

      // Now check if the update was successful
      res = await this.$frontend.rest("GET", "api/server/updates/status");
      this.restarting = false;
      if (!res.response.ok) {
        console.log("Update error: ", res.data.error_description);
        this.$store.dispatch("getUpdates");
        this.alert = res.data.error_description;
      } else {
        // Perform a refresh, the update might have activated plugins/modified the frontend
        location.reload(true);
      }
    }
  },
  created() {
    this.$store.dispatch("getUpdates");
  }
};
</script>