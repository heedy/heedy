<template>
  <h-card-page :title="'Update ' + source.name" :alert="alert">
    <v-container fluid grid-list-md>
      <v-layout row>
        <v-flex sm5 md4 xs12>
          <h-icon-editor
            ref="iconEditor"
            :image="source.icon"
            :colorHash="source.id"
            defaultIcon="timeline"
          ></h-icon-editor>
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
      <v-btn dark color="red" @click="del" :loading="loading">Delete</v-btn>
      <v-btn text @click="advanced = !advanced">
        <v-icon left>{{advanced? "expand_less":"expand_more"}}</v-icon>Advanced
      </v-btn>
      <v-spacer></v-spacer>
      <v-btn dark color="blue" @click="update" :loading="loading">Update</v-btn>
    </v-card-actions>
  </h-card-page>
</template>
<script>
export default {
  props: {
    source: Object
  },
  data: () => ({
    alert: "",
    modified: {},
    loading: false,
    advanced: false,
    scode: null,
    ssubtype: null,
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
    update: async function() {
      if (this.loading) return;

      this.loading = true;
      this.alert = "";

      if (this.$refs.iconEditor.hasImage()) {
        // We are in the image picker, and an image was chosen
        this.modified.icon = this.$refs.iconEditor.getImage();
      }

      let meta = {};

      if (this.advanced && this.scode != null) {
        try {
          var s = JSON.parse(this.scode);
        } catch {
          this.alert = "Could not parse schema";
          this.loading = false;
          return;
        }
        meta.schema = s;
      }
      if (this.ssubtype != null) {
        meta.subtype = this.ssubtype;
      }

      let mod = this.modified;

      if (Object.keys(meta).length > 0) {
        mod.meta = {
          ...this.source.meta,
          ...meta
        };
      }
      if (Object.keys(this.modified).length > 0) {
        console.log("UPDATING", mod);
        let result = await this.$app.api(
          "PATCH",
          `api/heedy/v1/sources/${this.source.id}`,
          mod
        );

        if (!result.response.ok) {
          this.alert = result.data.error_description;
          this.loading = false;
          return;
        }
        this.$store.dispatch("readSource", {
          id: this.source.id
        });
      }

      this.loading = false;
      this.$router.go(-1);
    },
    del: async function() {
      let s = this.source;
      if (
        confirm(
          `Are you sure you want to delete '${this.source.name}'? This deletes all associated data.`
        )
      ) {
        let res = await this.$app.api(
          "DELETE",
          `/api/heedy/v1/sources/${this.source.id}`
        );
        if (!res.response.ok) {
          this.alert = res.data.error_description;
        } else {
          this.alert = "";
          if (s.app != null) {
            this.$router.push(`/apps/${s.app}`);
          } else {
            this.$router.push(`/users/${s.owner}`);
          }
        }
      }
    }
  },
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
    code: {
      get() {
        if (this.scode != null) {
          return this.scode;
        }
        return JSON.stringify(this.source.meta.schema);
      },
      set(v) {
        this.scode = v;
      }
    },
    subtype: {
      get() {
        if (this.ssubtype != null) {
          return this.ssubtype;
        }
        return this.source.meta.subtype || "";
      },
      set(v) {
        this.ssubtype = v;
      }
    },
    description: {
      get() {
        if (this.modified.description !== undefined) {
          return this.modified.description;
        }
        return this.source.description;
      },
      set(v) {
        this.$app.vue.set(this.modified, "description", v);
      }
    },
    name: {
      get() {
        if (this.modified.name !== undefined) {
          return this.modified.name;
        }
        return this.source.name;
      },
      set(v) {
        this.$app.vue.set(this.modified, "name", v);
      }
    }
  }
};
</script>