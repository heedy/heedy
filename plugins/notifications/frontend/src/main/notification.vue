<template>
  <v-alert
    :type="
      n.type.length > 0 ? (n.type == 'toolbar' ? undefined : n.type) : 'info'
    "
    border="left"
    :colored-border="false"
    :dismissible="n.actions.length < 1 && n.dismissible"
    :dense="true"
    :outlined="true"
    prominent
    :elevation="1"
    :icon="false"
    :color="n.type == 'toolbar' ? 'rgba(0,0,0,.6)' : undefined"
    @input="del"
    style="background-color: #fdfdfd !important"
  >
    <v-row no-gutters>
      <v-col
        :cols="
          n.actions.length == 0 || actionCols.actions == 12
            ? 12
            : 12 - actionCols.actions
        "
        :style="titleStyle.col"
      >
        <h3 :style="titleStyle.h3">
          <router-link :to="linkpath" v-if="showlink">{{
            n.title
          }}</router-link>
          <span v-else>{{ n.title }}</span>
        </h3>
        <h-md
          v-if="description.length > 0"
          :value="description"
          style="padding-top: 5px"
        />
      </v-col>
      <v-col
        v-if="n.actions.length > 0"
        class="text-center"
        :cols="actionCols.actions"
      >
        <v-container :style="{ padding: 0, margin: 0 }" fluid>
          <v-row no-gutters>
            <v-col
              v-for="(v, i) in n.actions"
              :key="i"
              :cols="actionCols.buttons"
              style="padding-left: 2px; padding-right: 2px"
            >
              <v-tooltip
                bottom
                :disabled="v.tooltip === undefined || v.tooltip == ''"
              >
                <template v-slot:activator="{ on, attrs }">
                  <v-btn
                    outlined
                    v-bind="attrs"
                    v-on="on"
                    @click="runAction(v)"
                    :color="n.type.length > 0 ? n.type : 'info'"
                    style="width: 100%; margin: 2px"
                  >
                    <v-icon v-if="v.icon !== undefined && v.icon != ''" left>{{
                      v.icon
                    }}</v-icon>
                    {{ v.title }}
                  </v-btn>
                </template>
                <span>{{ v.tooltip }}</span>
              </v-tooltip>
            </v-col>
            <v-col
              :cols="actionCols.buttons"
              style="padding-left: 2px; padding-right: 2px"
            >
              <v-btn
                v-if="n.actions.length >= 1 && n.dismissible"
                outlined
                :color="n.type.length > 0 ? n.type : 'info'"
                style="width: 100%; margin: 2px"
                @click="del"
                ><v-icon left>cancel</v-icon>Dismiss</v-btn
              >
            </v-col>
          </v-row></v-container
        >
      </v-col>
    </v-row>
    <v-dialog
      v-if="dialog"
      :value="true"
      @input="cancelButton"
      max-width="800px"
    >
      <v-card>
        <v-card-text
          v-if="dialogSchema == null && actionDescription.length > 0"
        >
          <h-md :value="actionDescription" />
        </v-card-text>
        <v-form
          v-model="formValid"
          @submit="
            (e) => {
              e.preventDefault();
              linkTo(dialogAction);
            }
          "
        >
          <v-card-text v-if="dialogSchema != null">
            <v-progress-linear
              v-if="loading"
              :value="uploadPercent"
            ></v-progress-linear>
            <h-jsf v-else :schema="dialogSchema" v-model="actionForm" />
          </v-card-text>
          <v-card-text v-if="alert.length > 0">
            <v-alert text outlined color="deep-orange" icon="error_outline">{{
              alert
            }}</v-alert>
          </v-card-text>
          <v-card-actions>
            <v-btn text @click="cancelButton">
              {{ dialogAction.type == "md" ? "Dismiss" : "Cancel" }}
            </v-btn>
            <v-spacer />
            <v-btn
              v-if="dialogAction.href !== undefined && dialogAction.href != ''"
              :disabled="!formValid"
              color="primary"
              type="submit"
              :loading="loading"
            >
              {{ dialogAction.type == "link" ? dialogAction.title : "Submit" }}
            </v-btn>
          </v-card-actions>
        </v-form>
      </v-card>
    </v-dialog>
  </v-alert>
</template>
<script>
/*
There are several different cases for views
- If only actions, show just the action buttons
- If actions and title, but no description
  - If title is short, and not a lot of buttons, show buttons to side in a row
  - Otherwise have title, then buttons under it
- If description, show actions to the side, but show two columns of buttons if there are many
*/

function convertURL(href) {
  if (href.startsWith("/")) {
    href = location.href.split("#")[0] + href.substr(1);
  }
  return href;
}

export default {
  props: {
    n: Object,
    link: {
      type: Boolean,
      default: false,
    },
    small: {
      type: Boolean,
      default: false,
    },
    seen: {
      type: Boolean,
      default: false,
    },
  },
  data: () => ({
    dialog: false,
    dialogAction: {},
    actionForm: {},
    formValid: false,
    loading: false,
    uploadPercent: 0,
    xhr: null,
    alert: "",
  }),
  computed: {
    titleStyle() {
      if (this.description.length > 0) {
        return { col: "", h3: "padding-top:10px" };
      }
      if (this.n.actions.length == 0 || this.actionCols.actions == 12) {
        return { col: "", h3: "padding-top:0" };
      }
      return {
        col: "position: relative",
        h3: {
          position: "absolute",
          top: "50%",
          transform: "translate(0,-50%)",
        },
      };
    },
    description() {
      if (this.n.description === undefined) {
        return "";
      }
      return this.n.description;
    },
    actionDescription() {
      if (this.dialogAction.description === undefined) {
        return "";
      }
      return this.dialogAction.description;
    },
    showlink() {
      if (!this.link) return false;
      if (this.n.app !== undefined) return true;
      return false;
    },
    linkpath() {
      if (this.n.object !== undefined) return `/objects/${this.n.object}`;
      if (this.n.app !== undefined) return `/apps/${this.n.app}`;
      return `/users/${this.n.user}`;
    },
    dialogSchema() {
      if (
        this.dialogAction.form_schema === undefined ||
        Object.keys(this.dialogAction.form_schema).length == 0 ||
        this.dialogAction.form_schema.const !== undefined
      ) {
        return null;
      }
      if (this.dialogAction.form_schema.type !== undefined) {
        return this.dialogAction.form_schema;
      }
      let s = {
        type: "object",
        properties: {
          ...this.dialogAction.form_schema,
        },
      };
      if (s.properties.required !== undefined) {
        s.required = s.properties.required;
        delete s.properties.required;
      }
      if (
        this.dialogAction.description !== undefined &&
        this.dialogAction.description.length > 0
      ) {
        s.description = this.dialogAction.description;
      }
      return s;
    },
    actionCols() {
      if (this.$vuetify.breakpoint.xs) {
        // On phone screens we don't even worry about it
        return { buttons: 12, actions: 12 };
      }

      let actionButtons = this.n.actions.length;
      if (this.n.dismissible) {
        actionButtons += 1;
      }

      // Now let's count how many buttons we can show side by side if we do full width actions vs actions to the right
      let full_width_count = 6;
      if (this.$vuetify.breakpoint.sm) {
        full_width_count = 4;
      }

      if (this.description.length > 0) {
        if (this.description.length < 1000) {
          // If there is a description, then try doing at most 2 columns to the side
          if (actionButtons < 6) {
            return { buttons: 12, actions: 12 / full_width_count };
          }
          if (
            actionButtons <= 12 &&
            this.$vuetify.breakpoint.mdAndUp &&
            this.description.length < 500
          ) {
            return { buttons: 6, actions: (2 * 12) / full_width_count };
          }
        }

        // Otherwise, just add it at the bottom.
        return { buttons: 12 / full_width_count, actions: 12 };
      }

      // If we have just a title, then either show buttons on the side OR on bottom
      // with up to 3 buttons on the side
      if (actionButtons <= 3) {
        if (
          this.$vuetify.breakpoint.lgAndUp ||
          (this.$vuetify.breakpoint.mdAndUp && actionButtons <= 2) ||
          (this.$vuetify.breakpoint.smAndUp && actionButtons <= 1)
        ) {
          return {
            buttons: 12 / actionButtons,
            actions: (actionButtons * 12) / full_width_count,
          };
        }
      }

      // Otherwise, just add it at the bottom.
      return { buttons: 12 / full_width_count, actions: 12 };
    },
  },
  methods: {
    del(v) {
      let nq = { key: this.n.key };
      if (this.n.object !== undefined) {
        nq.object = this.n.object;
      } else if (this.n.app !== undefined) {
        nq.app = this.n.app;
      } else {
        nq.user = this.n.user;
      }
      this.$store.dispatch("deleteNotification", nq);
    },
    postForm: async function () {
      // We have 2 types of post. The first is json post, which is default, then there is form-data.
      this.loading = true;
      this.alert = "";

      let xhr = new XMLHttpRequest();
      xhr.upload.addEventListener(
        "progress",
        (evt) => {
          if (evt.lengthComputable) {
            this.uploadPercent = Math.floor((100 * evt.loaded) / evt.total);
          }
        },
        false
      );

      let endRequest = () => {
        this.xhr = null;
        this.uploadPercent = 0;
        this.loading = false;
      };
      xhr.addEventListener("load", (evt) => {
        if (evt.target.status != 200) {
          try {
            this.alert =
              "Failure: " + JSON.parse(evt.target.response).error_description;
          } catch {
            this.alert = "Failed to Submit";
          }
          endRequest();
          return;
        }
        endRequest();
        // Success!
        this.dialog = false;
        if (this.dialogAction.dismiss) {
          this.del(null);
        }
      });
      xhr.addEventListener("error", (evt) => {
        console.vlog("XHR ERROR", evt);
        endRequest();
        this.alert = "Upload failed";
      });
      xhr.addEventListener("abort", (evt) => {
        console.vlog("XHR ABORT", evt);
        endRequest();
      });

      this.xhr = xhr;
      let posturl = convertURL(this.dialogAction.href);
      console.vlog("Notification Action POST ", posturl);
      xhr.open("POST", posturl);

      if (this.dialogAction.type == "post/form-data") {
        let form = new FormData();
        Object.entries(this.actionForm).forEach(([k, v]) => {
          if (typeof v === "object") {
            if (this.dialogSchema.properties[k] !== undefined) {
              let s = this.dialogSchema.properties[k];
              if (s.contentMediaType !== undefined && v.data instanceof Blob) {
                // Add it as a file blob!
                console.vlog("Notification Action: File blob at key", k);
                form.append(k, v.data);
                return;
              }
            }
            form.append(k, JSON.stringify(v));
          } else {
            form.append(k, v);
          }
        });
        xhr.send(form);
      } else {
        xhr.setRequestHeader("Content-Type", "application/json");
        xhr.send(JSON.stringify(this.actionForm));
      }
    },
    blindPost: async function (v) {
      let data = null;
      if (v.form_schema !== undefined && v.form_schema.const !== undefined) {
        data = v.form_schema.const;
      }
      let post_type = v.type === "post/form-data" ? "form-data" : "json";
      let posturl = convertURL(v.href);
      console.vlog("Notification Action POST ", posturl);
      let res = await this.$frontend.rest(
        "POST",
        posturl,
        data,
        null,
        post_type
      );
      if (!res.response.ok) {
        this.alert = "Failed to Submit";
        this.dialog = true;
        return;
      }

      if (v.dismiss) {
        this.del(null);
      }
      this.dialog = false;
    },
    cancelButton(b) {
      if (this.xhr != null) {
        this.xhr.abort();
      }
      this.dialog = false;
      this.dialogAction = {};
    },
    linkTo(v) {
      if (this.loading) {
        return;
      }
      this.alert = "";

      if (v.type.startsWith("post")) {
        // If it is actively a form dialog, use the data input there.
        if (this.dialogSchema != null) {
          this.postForm();
          return;
        }

        // Otherwise, we do a blind post, using the const data if available, and nothing if not.
        this.blindPost(v);
        return;
      }

      let url = v.href;

      if (v.dismiss) {
        this.del(null);
      }
      if (url.startsWith("#")) {
        url = location.href.split("#")[0] + url;
        if (!v.new_window) {
          this.$router.push(v.href.substr(1));
          return;
        }
      } else if (v.href.startsWith("/")) {
        url = location.href.split("#")[0] + v.href.substr(1);
      }
      if (v.new_window) {
        window.open(url, "_blank");
      } else {
        location.href = url;
      }
    },
    runAction(v) {
      if (
        v.description !== undefined ||
        (v.form_schema !== undefined &&
          Object.keys(v.form_schema).length > 0 &&
          v.form_schema.const === undefined)
      ) {
        this.dialogAction = v;
        this.actionForm = {};
        this.dialog = true;
        this.alert = "";
        return;
      }
      this.linkTo(v);
    },
  },
  watch: {
    n(newN) {
      if (this.seen && !newN.seen) {
        let nq = { key: newN.key };
        if (newN.object !== undefined) {
          nq.object = newN.object;
        } else if (newN.app !== undefined) {
          nq.app = newN.app;
        } else {
          nq.user = newN.user;
        }
        this.$store.dispatch("updateNotification", {
          n: nq,
          u: { seen: true },
        });
      }
    },
  },
  created() {
    if (this.seen && !this.n.seen) {
      let nq = { key: this.n.key };
      if (this.n.object !== undefined) {
        nq.object = this.n.object;
      } else if (this.n.app !== undefined) {
        nq.app = this.n.app;
      } else {
        nq.user = this.n.user;
      }
      this.$store.dispatch("updateNotification", {
        n: nq,
        u: { seen: true },
      });
    }
  },
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