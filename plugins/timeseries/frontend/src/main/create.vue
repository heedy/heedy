<template>
  <h-card-page title="Create a new Timeseries" :alert="alert">
    <v-container fluid grid-list-md>
      <v-layout row>
        <v-flex sm5 md4 xs12>
          <h-icon-editor ref="iconEditor" image="timeline"></h-icon-editor>
        </v-flex>
        <v-flex sm7 md8 xs12>
          <v-container>
            <v-text-field
              label="Name"
              placeholder="My Timeseries"
              v-model="name"
            ></v-text-field>
            <v-text-field
              label="Description"
              placeholder="This timeseries holds my data"
              v-model="description"
            ></v-text-field>
            <h-tag-editor v-model="tags" />
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
        <v-icon left>{{ advanced ? "expand_less" : "expand_more" }}</v-icon
        >Advanced
      </v-btn>
      <v-spacer></v-spacer>
      <v-btn dark color="blue" @click="create" :loading="loading">Create</v-btn>
    </v-card-actions>
  </h-card-page>
</template>
<script>
export default {
  data: () => ({
    alert: "",
    advanced: false,
    loading: false,
    description: "",
    tags: "",
    code: "{}",
    name: "",
    cmOptions: {
      tabSize: 2,
      mode: "text/javascript",
    },
    schemaTypes: [
      {
        label: "Number",
        value: "number",
      },
      {
        label: "String",
        value: "string",
      },
      {
        label: "Other",
        value: "?",
      },
    ],
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
    },
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
    create: async function () {
      if (this.loading) return;

      this.loading = true;

      if (this.name == "") {
        this.alert = "Must fill in timeseries name";
        this.loading = false;
        return;
      }

      let toCreate = {
        name: this.name,
        type: "timeseries",
        description: this.description,
        tags: this.tags,
        meta: {},
        icon: this.$refs.iconEditor.getImage(),
      };
      if (this.advanced) {
        try {
          var s = JSON.parse(this.code);
        } catch {
          this.alert = "Could not parse schema";
          this.loading = false;
          return;
        }
        toCreate.meta.schema = s;
      }
      let result = await this.$frontend.rest("POST", `api/objects`, toCreate);

      if (!result.response.ok) {
        this.alert = result.data.error_description;
        this.loading = false;
        return;
      }
      // The result comes without the icon, let's set it correctly
      result.data.icon = toCreate.icon;

      this.$store.commit("setObject", result.data);
      this.loading = false;
      this.$router.replace({ path: `/objects/${result.data.id}` });
    },
  },
};
</script>
<style>
.CodeMirror {
  border: 1px solid #eee;
  height: auto;
}
</style>
