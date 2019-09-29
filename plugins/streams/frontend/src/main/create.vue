<template>
  <h-page-container noflex>
    <v-flex
      justify-center
      align-center
      text-center
      style="padding: 10px; padding-bottom: 20px;padding-top:20px;"
    >
      <h1 style="color:#1976d2;">Create a new Stream</h1>
    </v-flex>
    <v-flex>
      <v-card>
        <div style="padding: 10px; padding-bottom: 0;">
          <v-alert
            v-if="alert.length>0"
            text
            outlined
            color="deep-orange"
            icon="error_outline"
          >{{ alert }}</v-alert>
        </div>
        <v-container fluid grid-list-md>
          <v-layout row>
            <v-flex sm5 md4 xs12>
              <h-avatar-editor ref="avatarEditor" image="timeline"></h-avatar-editor>
            </v-flex>
            <v-flex sm7 md8 xs12>
              <v-container>
                <v-text-field label="Name" placeholder="My Stream" v-model="name"></v-text-field>
                <v-text-field
                  label="Description"
                  placeholder="This streams holds my data"
                  v-model="description"
                ></v-text-field>
                <v-text-field label="Subtype" placeholder></v-text-field>
              </v-container>
            </v-flex>
          </v-layout>
        </v-container>
        <v-container v-if="advanced">
          <v-row>
            <v-flex sm5 md4 xs12>
              <v-container>
                <v-radio-group :value="curRadio" @change="setRadio">
                  <v-radio
                    v-for="item in schemaTypes"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  ></v-radio>
                </v-radio-group>
              </v-container>
            </v-flex>
            <v-flex sm7 md8 xs12>
              <v-container>
                <h5>JSON Schema</h5>
                <codemirror v-model="code" :options="cmOptions"></codemirror>
              </v-container>
            </v-flex>
          </v-row>
        </v-container>

        <v-card-actions>
          <v-btn text @click="advanced = !advanced">
            <v-icon left>{{advanced? "expand_less":"expand_more"}}</v-icon>Advanced
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn dark color="blue" @click="create" :loading="loading">Create</v-btn>
        </v-card-actions>
      </v-card>
    </v-flex>
  </h-page-container>
</template>
<script>
import api from "../../api.mjs";

export default {
  data: () => ({
    alert: "",
    advanced: false,
    loading: false,
    description: "",
    subtype: "",
    code: "{}",
    name: "",
    cmOptions: {
      tabSize: 2,
      mode: "text/javascript"
    },
    schemaTypes: [
      {
        label: "Number",
        value: "number"
      },
      {
        label: "String",
        value: "string"
      },
      {
        label: "Other",
        value: "?"
      }
    ]
  }),
  computed: {
    curRadio() {
      try {
        let s = JSON.parse(this.code);
        for (let i = 0; i < this.schemaTypes.length; i++) {
          if (this.schemaTypes[i].value == s.type) {
            return s.type;
          }
        }
      } catch {}
      return "?";
    }
  },
  methods: {
    setRadio(v) {
      switch (v) {
        case "?":
          this.code = "{}";
          return;
        default:
          this.code = JSON.stringify({ type: v }, null, "  ");
      }
    },
    create: async function() {
      if (this.loading) return;

      this.loading = true;

      if (this.name == "") {
        this.alert = "Must fill in stream name";
        this.loading = false;
        return;
      }

      let toCreate = {
        name: this.name,
        type: "stream"
      };
      if (this.advanced) {
        try {
          var s = JSON.parse(this.code);
        } catch {
          this.alert = "Could not parse schema";
          this.loading = false;
          return;
        }
        toCreate.meta = {
          schema: s,
          subtype: this.subtype
        };
        toCreate.description = this.description;
        toCreate.avatar = this.$refs.avatarEditor.getImage();
      }
      let result = await api("POST", `api/heedy/v1/sources`, toCreate);

      if (!result.response.ok) {
        this.alert = result.data.error_description;
        this.loading = false;
        return;
      }
      // The result comes without the avatar, let's set it correctly
      result.data.avatar = toCreate.avatar;

      this.$store.commit("setSource", result.data);
      this.loading = false;
      this.$router.push({ path: `/sources/${result.data.id}` });
    }
  }
};
</script>
<style>
.CodeMirror {
  border: 1px solid #eee;
  height: auto;
}
</style>