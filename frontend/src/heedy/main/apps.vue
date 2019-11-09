<template>
  <h-page-container>
    <v-card>
      <v-card-title>
        <v-list-item two-line>
          <v-btn color="blue darken-2" dark fab right absolute @click.stop="addAppList">
            <v-icon>add</v-icon>
          </v-btn>
          <v-list-item-content>
            <v-list-item-title class="headline mb-1">Apps</v-list-item-title>
            <v-list-item-subtitle>Services and devices connected to heedy</v-list-item-subtitle>
          </v-list-item-content>
        </v-list-item>
      </v-card-title>
      <v-container fluid>
        <div v-if="loading" style="color: gray; text-align: center;">Loading...</div>
        <div
          v-else-if="apps.length==0"
          style="color: gray; text-align: center;"
        >You don't have any apps.</div>
        <v-row no-gutters v-else>
          <v-col v-for="c in apps" :key="c.id" cols="12" xs="12" sm="6" md="6" lg="4" xl="3">
            <v-card class="pa-2" outlined tile>
              <v-list-item two-line subheader :to="`/apps/${c.id}`">
                <v-list-item-avatar>
                  <h-icon :image="c.icon" :colorHash="c.id"></h-icon>
                </v-list-item-avatar>
                <v-list-item-content>
                  <v-list-item-title>{{ c.name }}</v-list-item-title>
                  <v-list-item-subtitle>{{ c.description }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
    <v-dialog v-model="dialog" max-width="1024">
      <v-card>
        <v-card-title>
          <v-list-item two-line>
            <v-list-item-content>
              <v-list-item-title class="headline mb-1">Add App</v-list-item-title>
              <v-list-item-subtitle>Add services provided by plugins, or create your own.</v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>
        </v-card-title>

        <v-card-text>
          <v-row no-gutters>
            <v-col cols="12" xs="12" sm="6" md="6" lg="4" xl="3">
              <v-card class="pa-2" outlined tile>
                <v-list-item two-line subheader :to="`/create/app`">
                  <v-list-item-avatar>
                    <h-icon image="settings_input_component" colorHash="a"></h-icon>
                  </v-list-item-avatar>
                  <v-list-item-content>
                    <v-list-item-title>Custom App</v-list-item-title>
                    <v-list-item-subtitle>Get an access token to use with your own software</v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
              </v-card>
            </v-col>
            <v-col
              v-for="pa in pluginApps"
              :key="pa.key"
              cols="12"
              xs="12"
              sm="6"
              md="6"
              lg="4"
              xl="3"
            >
              <v-card class="pa-2" outlined tile>
                <v-list-item two-line subheader @click="addApp(pa.key)">
                  <v-list-item-avatar>
                    <h-icon
                      :image="pa.icon"
                      :colorHash="pa.key"
                      defaultIcon="settings_input_component"
                    ></h-icon>
                  </v-list-item-avatar>
                  <v-list-item-content>
                    <v-list-item-title>{{ pa.name }}</v-list-item-title>
                    <v-list-item-subtitle>{{ pa.description }}</v-list-item-subtitle>
                  </v-list-item-content>
                </v-list-item>
              </v-card>
            </v-col>
          </v-row>
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="secondary" text @click="dialog = false">Cancel</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </h-page-container>
</template>
<script>
export default {
  data: () => ({
    dialog: false
  }),
  computed: {
    loading() {
      return this.$store.state.heedy.apps == null;
    },
    apps() {
      let c = this.$store.state.heedy.apps;

      return Object.keys(c).map(k => c[k]);
    },
    pluginApps() {
      let pa = this.$store.state.heedy.plugin_apps;
      if (pa === null) {
        return [];
      }

      let vals = Object.keys(pa).map(k => ({
        ...pa[k],
        key: k
      }));
      if (this.loading) {
        return vals;
      }
      let apps = this.apps;
      return vals.filter(v =>
        v.unique ? apps.filter(a => a.plugin == v.key).length == 0 : true
      );
    }
  },
  methods: {
    addAppList() {
      this.$store.dispatch("getPluginApps");
      this.dialog = true;
    },
    addApp: async function(appkey) {
      this.dialog = false;
      console.log("Creating", appkey);
      let result = await this.$app.api("POST", `api/heedy/v1/apps?icon=true`, {
        plugin: appkey
      });

      if (!result.response.ok) {
        this.$store.commit("alert", {
          type: "error",
          text: result.data.error_description
        });
        return;
      }
      this.$store.commit("setApp", result.data);
      this.$router.push({ path: `/apps/${result.data.id}` });
    }
  },
  created() {
    this.$store.dispatch("listApps");
  }
};
</script>