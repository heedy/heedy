<template>
  <h-object-updater :object="object" :meta="meta">
    <template v-slot:advanced>
      <v-text-field label="Subtype" v-model="subtype" placeholder></v-text-field>
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
    </template>
  </h-object-updater>
</template>
<script>
export default {
  props: {
    object: Object
  },
  data: () => ({
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
        return JSON.stringify(this.object.meta.schema);
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
        return this.object.meta.subtype || "";
      },
      set(v) {
        this.ssubtype = v;
      }
    },
    meta() {
      let meta = {};
      if (this.scode != null) {
        try {
          meta.schema = JSON.parse(this.scode);
        } catch {}
      }
      if (this.ssubtype != null) {
        meta.subtype = this.ssubtype;
      }
      return meta;
    }
  }
};
</script>
