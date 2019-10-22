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
      <v-dialog v-model="uploader" width="500" persistent>
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
            <v-checkbox :input-value="pi.active" @change="(v) => changeActive(pi.name,v)"></v-checkbox>
          </v-list-item-action>

          <v-list-item-content>
            <v-list-item-title>{{ pi.name }}</v-list-item-title>
            <v-list-item-subtitle>{{ pi.description }}</v-list-item-subtitle>
          </v-list-item-content>
          <v-list-item-action>
            <v-btn icon style="padding-right: 15px" @click="() => showDetails(pi)">
              <v-icon color="grey lighten-1">fas fa-info-circle</v-icon>
            </v-btn>
          </v-list-item-action>
          <v-list-item-avatar style="padding-right: 30px;">
            <h-icon :image="pi.icon" :colorHash="pi.name"></h-icon>
          </v-list-item-avatar>
        </v-list-item>
      </v-list>

      <v-flex row>
        <div class="flex-grow-1"></div>
        <v-btn color="primary" dark class="mb-2" @click="update">Update</v-btn>
      </v-flex>
    </div>
    <div
      v-else
      style="color: gray; text-align: center; padding: 1cm;"
    >You don't have any plugins installed.</div>
    <v-dialog v-model="dialog" max-width="1024px">
      <v-card>
        <v-card-title class="headline grey lighten-2" primary-title>
          <v-list-item two-line>
            <v-list-item-avatar style="padding-right: 30px;">
              <h-icon :image="dvalue.icon" :colorHash="dvalue.name"></h-icon>
            </v-list-item-avatar>
            <v-list-item-content>
              <v-list-item-title>{{ dvalue.name }}</v-list-item-title>
              <v-list-item-subtitle>{{ dvalue.description }}</v-list-item-subtitle>
            </v-list-item-content>
            <v-list-item-action>
              <v-checkbox
                label="Enabled"
                :input-value="dvalue.active"
                @change="(v) => changeActive(dvalue.name,v)"
              ></v-checkbox>
            </v-list-item-action>
          </v-list-item>
        </v-card-title>

        <v-card-text style="padding-top: 20px;">
          <span v-html="getMD"></span>
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions>
          <h5>
            {{ dvalue.version }} - {{ dvalue.license }}
            <div v-if="dvalue.homepage.length > 0">
              -
              <a :href="dvalue.homepage">homepage</a>
            </div>
          </h5>
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
    dvalue: {
      name: "",
      description: "",
      readme: "",
      license: "",
      homepage: "",
      version: "",
      icon: ""
    }
  }),
  computed: {
    pluginItems() {
      console.log("GET ITEMS", this.plugins, this.active);
      let obj = Object.keys(this.plugins).map(k => ({
        name: k,
        description:
          this.plugins[k].description !== undefined
            ? this.plugins[k].description
            : "",
        version:
          this.plugins[k].version !== undefined
            ? this.plugins[k].version
            : "v???",
        readme:
          this.plugins[k].readme !== undefined ? this.plugins[k].readme : "",
        license:
          this.plugins[k].license !== undefined
            ? this.plugins[k].license
            : "unlicensed",
        homepage:
          this.plugins[k].homepage !== undefined
            ? this.plugins[k].homepage
            : "",
        icon:
          this.plugins[k].icon !== undefined
            ? this.plugins[k].icon
            : "fas fa-puzzle-piece",
        active: this.active.includes(k)
      }));

      console.log(obj);

      return obj;
    },
    getMD() {
      return md.render(this.dvalue.readme);
    }
  },
  methods: {
    showDetails(p) {
      this.dvalue = p;
      this.dialog = true;
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
      xhr.open("POST", "api/heedy/v1/server/updates/plugins");
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
      let res = await this.$app.api(
        "PATCH",
        "api/heedy/v1/server/updates/config",
        { plugins: this.active }
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
      this.$app.api("GET", "api/heedy/v1/server/updates/plugins").then(res => {
        if (!res.response.ok) {
          this.alert = res.data.error_description;
          this.plugins = {};
          return;
        }
        console.log("plugins", res.data);
        this.plugins = res.data;
      });
      this.$app.api("GET", "api/heedy/v1/server/updates/config").then(res => {
        if (!res.response.ok) {
          this.alert = res.data.error_description;
          this.active = [];
          return;
        }
        console.log("active", res.data.plugins);
        this.active = res.data.plugins;
      });
    }
  },
  created() {
    this.reload();
  }
};
</script>
<style>
p {
  padding-top: 15px;
}
h1 {
  padding-top: 15px;
}
h2 {
  padding-top: 15px;
}
h3 {
  padding-top: 15px;
}
h4 {
  padding-top: 15px;
}
</style>