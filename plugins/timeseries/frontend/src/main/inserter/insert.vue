<template>
  <v-form @submit="insert" v-model="formValid">
    <div ref="jsform" v-if="!loading">
      <div
        v-if="customTimestamp"
        style="
          width: 100%;
          text-align: center;
          background-color: #e8f4f8;
          border-radius: 3px;
          padding: 10px;
          padding-bottom: 5px;
        "
      >
        <vc-date-picker v-model="date" mode="dateTime">
          <template v-slot="{ inputValue, inputEvents }">
            <v-text-field
              label="Timestamp"
              :value="inputValue"
              v-on="inputEvents"
            />
          </template>
        </vc-date-picker>
      </div>
      <v-jsf
        :schema="schema"
        :options="options"
        v-model="modified"
        class="markdownview"
      >
        <template
          v-for="ins in inserters"
          :slot="`custom-` + ins.k"
          slot-scope="{ value, label, on }"
        >
          <component
            :key="ins.k"
            :is="ins.v"
            :value="value"
            v-on="on"
            :label="label"
          />
        </template>
      </v-jsf>
    </div>
    <!-- https://github.com/koumoul-dev/vuetify-jsonschema-form/issues/21 -->
    <div
      v-else
      :style="{ height, textAlign: 'center', display: 'flex', margin: 'auto' }"
    >
      <h4 style="margin: auto">Inserting...</h4>
    </div>
    <div class="text-center" style="width: 100%">
      <v-btn dark color="info" type="submit" :loading="loading">Insert</v-btn>
    </div>
  </v-form>
</template>
<script>
import moment from "../../../dist/moment.mjs";
import { md } from "../../../dist/markdown-it.mjs";
export default {
  props: {
    object: Object,
    customTimestamp: {
      type: Boolean,
      default: false,
    },
  },
  data: () => ({
    formValid: false,
    loading: false,
    height: "20px",
    options: {
      markdown: (r) => {
        if (r === undefined || r == null || r == "") {
          return null;
        }
        return md.render(r);
      },
    },
    date: new Date(),
    modified: {},
  }),
  computed: {
    inserters() {
      let d = this.$store.state.timeseries.customInserters;
      return Object.keys(d).map((k) => ({ k: k, v: d[k] }));
    },
    schema() {
      if (
        this.object.meta.schema.type !== undefined &&
        this.object.meta.schema.type == "object"
      ) {
        return {
          type: "object",
          properties: {
            data: this.object.meta.schema,
          },
          required: ["data"],
        };
      }
      return {
        type: "object",
        properties: {
          data: {
            title: " ",
            ...this.object.meta.schema,
          },
        },
        required: ["data"],
      };
    },
  },
  watch: {
    customTimestamp(ts) {
      if (ts) {
        this.date = new Date();
      }
    },
  },
  methods: {
    insert: async function (event) {
      event.preventDefault();
      if (this.loading) return;
      if (!this.formValid) {
        return;
      }

      this.height = this.$refs.jsform.clientHeight + "px";

      this.loading = true;

      let ts = moment().unix();
      if (this.customTimestamp) {
        if (this.date == null) {
          console.error("Invalid custom timestamp");
          this.loading = false;
          return;
        }
        ts = moment(this.date).unix();
      }
      console.vlog("Inserting datapoint:", ts, this.modified.data);
      let res = await this.$frontend.rest(
        "POST",
        `api/objects/${encodeURIComponent(this.object.id)}/timeseries`,
        [{ t: ts, d: this.modified.data }]
      );

      if (!res.response.ok) {
        console.error(res);
        this.loading = false;
        return;
      }
      this.modified = { data: null };
      this.loading = false;
      this.$emit("inserted");
    },
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