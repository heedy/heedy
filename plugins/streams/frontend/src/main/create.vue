<template>
  <v-content>
    <v-container>
      <v-layout column>
        <v-flex justify-center align-center text-center style="color: lightgrey; padding: 10px">
          <h1>Create Stream</h1>
        </v-flex>
        <v-flex>
          <v-card>
            <v-card-title>
              <v-text-field label="Full Name" placeholder="My Stream" v-model="fullname"></v-text-field>
              <v-text-field label="Stream Name" placeholder="mystream" v-model="name"></v-text-field>
            </v-card-title>
            <v-card-text class="text--primary" v-if="advanced">
              <v-container fluid grid-list-md>
                <v-layout row>
                  <v-flex sm5 md4 xs12>
                    <h-avatar-editor ref="avatarEditor" image="timeline"></h-avatar-editor>
                  </v-flex>
                  <v-flex sm7 md8 xs12>
                    <v-container>
                      <v-textarea solo label="Stream Description" v-model="description"></v-textarea>
                      <v-text-field solo label="Subtype" placeholder></v-text-field>
                    </v-container>
                  </v-flex>
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
                </v-layout>
              </v-container>
            </v-card-text>
            <v-card-actions>
              <v-btn text @click="advanced = !advanced">
                <v-icon left>{{advanced? "expand_less":"expand_more"}}</v-icon>Advanced
              </v-btn>
              <v-spacer></v-spacer>
              <v-btn dark color="blue" @click="create" :loading="loading">Create</v-btn>
            </v-card-actions>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </v-content>
</template>
<script>
import api from "../../api.mjs";

export default {
  data: () => ({
    advanced: false,
    loading: false,
    description: "",
    subtype: "",
    code: "{}",
    name: "",
    fullname: "",
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

      if (this.name == "" || this.fullname == "") {
        this.$store.dispatch("errnotify", {
          error: "BAD_REQUEST",
          error_description: "Must fill in stream name and fullname"
        });
        this.loading = false;
        return;
      }

      let toCreate = {
        name: this.name,
        fullname: this.fullname,
        type: "stream"
      };
      if (this.advanced) {
        try {
          var s = JSON.parse(this.code);
        } catch {
          this.$store.dispatch("errnotify", {
            error: "BAD_REQUEST",
            error_description: "Could not parse schema"
          });
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
      let result = await api("POST", `api/heedy/v1/source`, toCreate);

      if (!result.response.ok) {
        this.$store.dispatch("errnotify", result.data);
        this.loading = false;
        return;
      }
      this.$store.commit("setSource", result.data);
      this.$router.push({ path: `/source/${result.data.id}/stream` });
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