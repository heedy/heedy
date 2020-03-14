<template>
  <div>
    <v-flex v-if="alert.length>0">
      <div style="padding: 10px; padding-bottom: 0;">
        <v-alert text outlined color="deep-orange" icon="error_outline">{{ alert }}</v-alert>
      </div>
    </v-flex>
    <v-toolbar flat color="white">
      <v-toolbar-title>Installed Plugins</v-toolbar-title>
      <v-spacer />
      <v-dialog v-model="uploader" max-width="500" persistent>
        <template v-slot:activator="{ on }">
          <v-btn outlined class="mb-2" v-on="on">Upload</v-btn>
        </template>
        <v-card>
          <v-card-title>Upload Plugin</v-card-title>
          <v-card-text>
            <v-file-input
              v-model="file"
              v-if="!uploading"
              show-size
              accept="application/zip"
              label="Zipped Plugin Folder"
            ></v-file-input>
            <v-progress-linear v-else :value="uploadPercent"></v-progress-linear>
          </v-card-text>
          <v-card-actions>
            <v-btn text @click="cancelUpload">Cancel</v-btn>
            <v-spacer></v-spacer>
            <v-btn color="primary" text @click="upload">Upload</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-toolbar>
    <div v-if="pluginItems.length > 0">
      <v-list>
        <v-list-item v-for="pi in pluginItems" :key="pi.name" two-line>
          <v-list-item-action>
            <v-tooltip bottom>
              <template v-slot:activator="{ on }">
                <v-checkbox
                  v-on="on"
                  :input-value="isActive(pi.name)"
                  @change="(v) => changeActive(pi.name,v)"
                ></v-checkbox>
              </template>
              <span>Enable/Disable Plugin</span>
            </v-tooltip>
          </v-list-item-action>

          <v-list-item-content>
            <v-list-item-title>{{ pi.name }}</v-list-item-title>
            <v-list-item-subtitle>{{ pi.description }}</v-list-item-subtitle>
          </v-list-item-content>
          <v-list-item-action>
            <v-tooltip bottom>
              <template v-slot:activator="{ on }">
                <v-btn icon v-on="on" @click="() => showDetails(pi)">
                  <v-icon color="grey">fas fa-info-circle</v-icon>
                </v-btn>
              </template>
              <span>Plugin Info</span>
            </v-tooltip>
          </v-list-item-action>
          <v-list-item-avatar>
            <v-tooltip bottom>
              <template v-slot:activator="{ on }">
                <v-btn icon v-on="on" @click="() => deletePlugin(pi)">
                  <v-icon color="grey">fas fa-trash</v-icon>
                </v-btn>
              </template>
              <span>Delete Plugin</span>
            </v-tooltip>
          </v-list-item-avatar>
        </v-list-item>
      </v-list>
    </div>
    <div
      v-else
      style="color: gray; text-align: center; padding: 1cm;"
    >You don't have any plugins installed.</div>
    <v-dialog v-if="plugins[dvalue]!==undefined" v-model="dialog" max-width="1024px">
      <v-card>
        <v-card-title class="headline grey lighten-2" primary-title>
          <v-list-item two-line style="overflow:hidden;">
            <v-list-item-avatar>
              <h-icon :image="plugins[dvalue].icon" :colorHash="plugins[dvalue].name"></h-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>{{ plugins[dvalue].name }}</v-list-item-title>
              <v-list-item-subtitle>{{ plugins[dvalue].description }}</v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-action v-if="!$vuetify.breakpoint.sm && !$vuetify.breakpoint.xs">
              <v-checkbox
                label="Enabled"
                :input-value="isActive(dvalue)"
                @change="(v) => changeActive(dvalue,v)"
              ></v-checkbox>
            </v-list-item-action>
            <v-list-item-action v-if="!$vuetify.breakpoint.sm && !$vuetify.breakpoint.xs">
              <v-tooltip bottom>
                <template v-slot:activator="{ on }">
                  <v-btn icon v-on="on" @click="() => {dialog=false;deletePlugin(plugins[dvalue])}">
                    <v-icon>fas fa-trash</v-icon>
                  </v-btn>
                </template>
                <span>Delete Plugin</span>
              </v-tooltip>
            </v-list-item-action>
          </v-list-item>
        </v-card-title>

        <v-card-text style="padding-top: 20px;">
          <v-container fluid v-if="plugins[dvalue].readme===undefined">
            <v-layout justify-center align-center>
              <v-flex text-center>
                <h1>Loading...</h1>
              </v-flex>
            </v-layout>
          </v-container>
          <span v-else v-html="getMD" class="markdownview"></span>
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions>
          <h5 v-if="!$vuetify.breakpoint.sm && !$vuetify.breakpoint.xs">
            {{ plugins[dvalue].version }} - {{ plugins[dvalue].license }}
            <div v-if="plugins[dvalue].homepage.length > 0">
              -
              <a :href="plugins[dvalue].homepage">homepage</a>
            </div>
          </h5>
          <v-checkbox
            v-else
            label="Enabled"
            :input-value="isActive(dvalue)"
            @change="(v) => changeActive(dvalue,v)"
          ></v-checkbox>
          <v-spacer></v-spacer>
          <v-btn color="primary" text @click="dialog = false">ok</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>
<script>
import { md } from "../../../dist/markdown-it.mjs";

export default {
  data: () => ({
    plugins: {},
    active: [],
    alert: "",
    dialog: false,
    uploader: false,
    file: null,
    uploading: false,
    uploadPercent: 0,
    xhr: null,
    dvalue: ""
  }),
  computed: {
    pluginItems() {
      let pvals = Object.values(this.plugins);
      let o = this.$store.state.heedy.updates.options;
      if (o == null || o.deleted == null) {
        return pvals;
      }
      return pvals.filter(v => !o.deleted.includes(v.name));
    },
    getMD() {
      return md.render(this.plugins[this.dvalue].readme);
    }
  },
  methods: {
    showDetails(p) {
      this.dvalue = p.name;
      this.dialog = true;
      this.getReadme(p.name);
    },
    deletePlugin: async function(p) {
      if (
        confirm(
          `Are you sure you want to delete plugin '${p.name}'? You can disable it instead.`
        )
      ) {
        let o = this.$store.state.heedy.updates.options;
        if (o == null) {
          o = {
            backup: true,
            deleted: []
          };
        }
        o = {
          ...o,
          deleted: [...o.deleted, p.name]
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
      }
    },
    changeActive(pname, v) {
      console.log(pname, v);
      let pindex = this.active.indexOf(pname);
      if (pindex > -1) {
        if (!v) {
          this.active.splice(pindex, 1);
        }
      } else {
        if (v) {
          this.active.push(pname);
        }
      }
      this.update();
    },
    upload: async function() {
      if (this.uploading) {
        console.log("Already uploading.");
        return;
      }
      console.log(this.file);
      if (this.file == null) {
        console.log("no file selected");
        return;
      }
      this.alert = "";
      this.uploading = true;
      this.uploadPercent = 0;

      let form = new FormData();

      form.append("zipfile", this.file);

      var xhr = new XMLHttpRequest();
      xhr.upload.addEventListener(
        "progress",
        evt => {
          if (evt.lengthComputable) {
            this.uploadPercent = Math.floor((100 * evt.loaded) / evt.total);
          }
        },
        false
      );
      let endRequest = () => {
        this.file = null;
        this.uploading = false;
        this.uploader = false;
        this.xhr = null;
        this.$store.dispatch("getUpdates");
      };
      xhr.addEventListener("load", evt => {
        if (evt.target.status != 200) {
          try {
            this.alert =
              "Upload failed: " +
              JSON.parse(evt.target.response).error_description;
          } catch {
            this.alert = "Upload failed";
          }

          return;
        }
        endRequest();
        this.reload();
      });
      xhr.addEventListener("error", evt => {
        console.log("ERROR", evt);
        endRequest();
        this.alert = "Upload failed";
      });
      xhr.addEventListener("abort", evt => {
        console.log("ABORT", evt);
        endRequest();
      });
      xhr.open("POST", "api/server/updates/plugins");
      xhr.send(form);
      this.xhr = xhr;
    },
    cancelUpload() {
      if (this.xhr != null) {
        this.xhr.abort();
      } else {
        this.uploading = false;
        this.uploader = false;
      }
    },
    update: async function() {
      console.log(this.active);
      let res = await this.$frontend.rest(
        "PATCH",
        "api/server/updates/config",
        {
          active_plugins: this.active
        }
      );
      if (!res.response.ok) {
        this.alert = res.data.error_description;
        return;
      }
      this.alert = "";
      this.$store.dispatch("getUpdates");
      this.reload();
    },
    reload: async function() {
      this.$frontend.rest("GET", "api/server/updates/plugins").then(res => {
        if (!res.response.ok) {
          this.alert = res.data.error_description;
          this.plugins = {};
          return;
        }
        console.log("plugins", res.data);

        let plugineer = {};
        Object.keys(res.data).map(k => {
          plugineer[k] = {
            name: k,
            description:
              res.data[k].description !== undefined
                ? res.data[k].description
                : "",
            version:
              res.data[k].version !== undefined ? res.data[k].version : "v???",
            license:
              res.data[k].license !== undefined
                ? res.data[k].license
                : "unlicensed",
            homepage:
              res.data[k].homepage !== undefined ? res.data[k].homepage : "",
            icon:
              res.data[k].icon !== undefined
                ? res.data[k].icon
                : "fas fa-puzzle-piece"
          };
        });

        this.plugins = plugineer;
      });
      this.$frontend.rest("GET", "api/server/updates/config").then(res => {
        if (!res.response.ok) {
          this.alert = res.data.error_description;
          this.active = [];
          return;
        }
        console.log("active", res.data.active_plugins);
        this.active = res.data.active_plugins;
      });
    },
    isActive(k) {
      return this.active.includes(k);
    },
    getReadme: async function(pname) {
      console.log("Getting readme for", pname);
      let setreadme = r => {
        this.plugins[pname] = {
          ...this.plugins[pname],
          readme: r
        };
      };
      try {
      } catch (err) {
        setreadme("Error getting plugin readme.");
        return;
      }
      let res = await fetch(`api/server/updates/plugins/${pname}/README.md`, {
        method: "GET",
        credentials: "include",
        redirect: "follow"
      });
      console.log("README", res);
      if (!res.ok) {
        // this.alert = res.data.error_description;
        console.log("Plugin has no README");
        setreadme("This plugin has no README.md");
        return;
      }
      setreadme(await res.text());
    }
  },

  created() {
    this.reload();
  }
};
</script>
<style>
.markdownview p {
  padding-top: 15px;
}
.markdownview h1 {
  padding-top: 15px;
}
.markdownview h2 {
  padding-top: 15px;
}
.markdownview h3 {
  padding-top: 15px;
}
.markdownview h4 {
  padding-top: 15px;
}
.markdownview img {
  max-width: 100%;
}
</style>