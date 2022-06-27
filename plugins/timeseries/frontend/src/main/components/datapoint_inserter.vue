<template>
  <v-form @submit="insert" v-model="formValid">
    <div ref="jsform" v-if="!loading">
      <v-row
        v-if="customTimestamp"
        style="
          text-align: center;
          background-color: #e8f4f8;
          border-radius: 3px;
          margin-bottom: 15px;
        "
      >
        <v-col
          cols="12"
          xs="12"
          :sm="editDuration ? 6 : 12"
          style="padding-bottom: 0"
        >
          <div
            :style="
              editDuration
                ? ''
                : 'display: grid; grid-template-columns: auto 50px'
            "
          >
            <vc-date-picker v-model="date" mode="dateTime">
              <template #default="{ inputValue, inputEvents, updateValue }">
                <v-text-field
                  label="Timestamp"
                  :value="inputValue"
                  @focus="inputEvents.focusin($event)"
                  @blur="inputEvents.focusout($event)"
                  @click="inputEvents.click"
                  @input="
                    updateValue($event, {
                      formatInput: false,
                      hidePopover: false,
                      debounce: 300,
                    })
                  "
                  @change="
                    updateValue($event, {
                      formatInput: true,
                      hidePopover: false,
                    })
                  "
                />
              </template>
            </vc-date-picker>
            <v-tooltip bottom v-if="!editDuration">
              <template v-slot:activator="{ on, attrs }">
                <v-btn
                  style="place-self: center"
                  icon
                  v-on="on"
                  v-bind="attrs"
                  @click="showDuration = true"
                >
                  <v-icon>hourglass_empty</v-icon>
                </v-btn>
              </template>
              <span>Edit Duration</span>
            </v-tooltip>
          </div>
        </v-col>
        <v-col
          cols="12"
          xs="12"
          sm="6"
          v-if="editDuration"
          style="padding-bottom: 0"
        >
          <h-duration-editor v-model="duration" allowNegative />
        </v-col>
      </v-row>
      <h-jsf :schema="schema" :options="options" v-model="modified" />
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
import {getUnixTime} from "../../../dist/date-fns.mjs";
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
    duration: 0,
    showDuration: false,
  }),
  computed: {
    editDuration() {
      return this.showDuration || this.duration != 0;
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

      let dp = { t: getUnixTime(new Date()), d: this.modified.data };
      if (this.customTimestamp) {
        if (this.date == null) {
          console.error("Invalid custom timestamp");
          this.loading = false;
          return;
        }
        dp.t = getUnixTime(this.date);
        if (isFinite(this.duration) && this.duration != 0) {
          if (this.duration < 0) {
            dp.t += this.duration;
            dp["dt"] = -this.duration;
          } else {
            dp["dt"] = this.duration;
          }
        }
      }
      console.vlog("Inserting datapoint:", dp);
      let res = await this.$frontend.rest(
        "POST",
        `api/objects/${encodeURIComponent(this.object.id)}/timeseries`,
        [dp]
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