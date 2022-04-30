<template>
<div>
    <v-toolbar flat color="white">
      <v-toolbar-title>Heedy {{ version }}</v-toolbar-title>
      <v-spacer />
      
    </v-toolbar>
    <v-card-text>
        <v-row no-gutters>
            <v-col
              style="padding-left: 2px; padding-right: 2px"
            >
                <v-btn
                    outlined
                    @click="setBackup(false)"
                    style="width: 100%; margin: 2px"
                  >
                    Restart Server
                  </v-btn>
            </v-col>
            <v-col
              style="padding-left: 2px; padding-right: 2px"
            >
                <v-btn
                    outlined
                    @click="setBackup(true)"
                    style="width: 100%; margin: 2px"
                  >
                    Backup Database
                  </v-btn>
            </v-col>
            <v-col
              style="padding-left: 2px; padding-right: 2px"
            >
                <v-btn
                    outlined
                    @click="revert=true"
                    style="width: 100%; margin: 2px"
                  >
                    Revert from Backup
                  </v-btn>
            </v-col>
            </v-row>
    </v-card-text>
    <v-flex v-if="alert.length > 0">
      <div style="padding: 10px; padding-bottom: 0">
        <v-alert text outlined color="deep-orange" icon="error_outline">{{
          alert
        }}</v-alert>
      </div>
    </v-flex>
    <v-flex v-if="revert">
      <div style="padding: 10px; padding-bottom: 0">
        <v-alert outlined type="warning" prominent border="left">
          <v-row align="center">
            <v-col class="grow">
                <b>You might lose all changes since the last backup. Continue?</b>
                <ul>
                    <li>Any plugins or configuration changes applied will be reverted.</li>
                    <li>If the database was part of the backup, all data added since then will be lost.</li>
                </ul>
            </v-col>
            <v-col class="shrink">
              <v-btn color="warning" style="width: 100%" outlined @click="runrevert"
                >Revert</v-btn
              >
              <v-btn
                color="warning"
                style="width: 100%"
                outlined
                @click="revert=false"
                >Cancel</v-btn
              >
            </v-col>
          </v-row>
        </v-alert>
      </div>
      <v-dialog v-model="reverting" persistent max-width="500px">
        <v-card>
          <v-card-title>Reverting...</v-card-title>
          <v-card-text>
            Please be patient while the server restarts. This may take several
            minutes if recovering a database backup or changing a plugin. This page will automatically reload once the
            server has restarted.
          </v-card-text>
        </v-card>
      </v-dialog>
    </v-flex>
    </div>
</template>
<script>
export default {
    data:() => ({
        alert: "",
        revert: false,
        reverting: false,
    }),
    computed: {
        version() {
            return this.$store.state.app.info.version;
        }
    },
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
    runrevert: async function () {
        this.alert="";
      this.$frontend.websocket.disable(); // Shut down the websocket
      let res = await this.$frontend.rest("POST", "api/server/restart", {
        revert: true,
      });
      if (!res.response.ok) {
        if (res.response.status !== undefined) {
          console.verror("Revert error: ", res.data.error_description);
          this.alert = res.data.error_description;
          this.revert = false;
          this.$frontend.websocket.enable();
          return;
        }
      }

      this.reverting = true;

      function sleep(ms) {
        return new Promise((resolve) => setTimeout(resolve, ms));
      }
      await sleep(1000);

      res = await this.$frontend.rest("GET", "api/server/version");
      while (!res.response.ok) {
        await sleep(1000);
        res = await this.$frontend.rest("GET", "api/server/version");
      }

      // If there is a service worker, update it to a new version if necessary before refreshing
        if ("serviceWorker" in navigator) {
          let registration = await navigator.serviceWorker.getRegistration(
            "/service-worker.js"
          );
          if (registration !== undefined) {
            registration = await registration.update();
            if (registration !== undefined) {
              let newWorker = registration.installing;
              if (newWorker != null) {
                console.vlog("Waiting for ServiceWorker to install");
                await Promise.race([
                  sleep(5000),
                  new Promise((resolve) => {
                    newWorker.addEventListener("statechange", () => {
                      if (newWorker.state === "installed") {
                        resolve();
                      }
                    });
                  }),
                ]);
              }
            }
          }
        }
        console.vlog("Reloading");
        // Perform a refresh, the update might have activated plugins/modified the frontend
        location.reload(true);
    }
    }
}
</script>