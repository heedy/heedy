<template>
  <v-dialog :value="true" @input="cancelEdit" max-width="600px">
    <v-card>
      <v-card-title>
        <h-icon
          :image="object.icon"
          :defaultIcon="defaultIcon"
          :colorHash="object.id"
          :size="30"
        />
        <span style="padding-left: 10px">{{ object.name }}</span>
        <v-spacer />
        <span style="color: gray; font-size: 12px">Editing Datapoint</span>
      </v-card-title>
      <v-alert
        v-if="alert.length > 0"
        text
        outlined
        color="deep-orange"
        icon="error_outline"
        style="margin: 10px; margin-bottom: 0"
        >{{ alert }}</v-alert
      >
      <v-form @submit="insert" v-model="formValid">
        <v-card-text ref="jsform" v-if="!loading">
          <v-row
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
                <vc-date-picker
                  v-model="timestamp"
                  mode="dateTime"
                  :popover="{ positionFixed: true }"
                >
                  <template v-slot="{ inputValue, inputEvents, updateValue }">
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
          <h-jsf :schema="schema" :options="options" v-model="newdata" />
        </v-card-text>
        <v-card-text
          v-else
          :style="{
            height,
            textAlign: 'center',
            display: 'flex',
            margin: 'auto',
          }"
        >
          <h4 style="margin: auto">Inserting...</h4>
        </v-card-text>
        <v-card-actions>
          <v-btn dark color="red" @click="deleter" :loading="loading">
            Delete
          </v-btn>
          <v-spacer />
          <v-btn text @click="cancelEdit"> Cancel </v-btn>
          <v-btn color="primary" type="submit" :loading="loading"> Save </v-btn>
        </v-card-actions>
      </v-form>
    </v-card>
  </v-dialog>
</template>
<script>
import { md } from "../../../../dist/markdown-it.mjs";

export default {
  props: {
    object: Object,
    value: Boolean,
    datapoint: Object,
  },
  data() {
    return {
      alert: "",
      loading: false,
      height: "20px",
      formValid: false,
      options: {
        markdown: (r) => {
          if (r === undefined || r == null || r == "") {
            return null;
          }
          return md.render(r);
        },
      },
      timestamp: new Date(this.datapoint.t * 1000),
      timestamp_changes: 0,
      newdata: { data: JSON.parse(JSON.stringify(this.datapoint.d)) },
      data_changes: 0,
      duration: this.datapoint.dt || 0,
      duration_changes: 0,
      showDuration: false,
    };
  },
  methods: {
    cancelEdit() {
      this.$emit("input", false);
    },
    insert: async function (event) {
      event.preventDefault();
      if (this.loading) return;
      if (!this.formValid) {
        return;
      }
      if (
        this.timestamp_changes == 0 &&
        this.data_changes == 0 &&
        this.duration_changes == 0
      ) {
        this.cancelEdit();
        return;
      }

      this.height = this.$refs.jsform.clientHeight + "px";
      this.loading = true;

      let newdp = JSON.parse(JSON.stringify(this.datapoint));
      if (this.timestamp_changes > 0) {
        newdp.t = this.timestamp.getTime() / 1000;
      }
      if (this.duration_changes > 0) {
        if (this.duration < 0) {
          newdp.t += this.duration;
          newdp.dt = -this.duration;
        } else {
          newdp.dt = this.duration;
          if (newdp.dt == 0) {
            delete newdp.dt;
          }
        }
      }
      if (this.data_changes > 0) {
        newdp.d = this.newdata.data;
      }

      console.vlog("Updating Datapoint", this.datapoint, newdp);

      if (this.timestamp_changes == 0 && this.duration_changes == 0) {
        // Since only the data changed, we can just update the datapoint
        let res = await this.$frontend.rest(
          "POST",
          `api/objects/${encodeURIComponent(this.object.id)}/timeseries`,
          [newdp]
        );
        if (!res.response.ok) {
          console.error(res);
          this.alert = res.data.error_description;
          this.loading = false;
          return;
        }
        this.loading = false;
        this.cancelEdit();
        return;
      }
      // If there were changes in timestamp/duration, we first delete
      // the old datapoint and then reinsert the new one.
      // This is done for durations to make sure that we don't overlap with existing data.
      let res = await this.$frontend.rest(
        "DELETE",
        `api/objects/${encodeURIComponent(this.object.id)}/timeseries`,
        { t: this.datapoint.t }
      );
      if (!res.response.ok) {
        console.error(res);
        this.alert = res.data.error_description;
        this.loading = false;
        return;
      }

      res = await this.$frontend.rest(
        "POST",
        `api/objects/${encodeURIComponent(this.object.id)}/timeseries`,
        [newdp],
        { method: "insert" }
      );

      if (!res.response.ok) {
        console.error(res);
        this.alert = res.data.error_description;
        // Now try to reinsert the old datapoint that we previously deleted.
        await this.$frontend.rest(
          "POST",
          `api/objects/${encodeURIComponent(this.object.id)}/timeseries`,
          [this.datapoint],
          { method: "insert" }
        );
        this.loading = false;
        return;
      }
      this.loading = false;
      this.cancelEdit();
    },
    deleter: async function () {
      if (this.loading) return;

      // Check to make sure the user wants to delete the datapoint
      if (
        !confirm(
          "Are you sure you want to delete this datapoint? This cannot be undone."
        )
      ) {
        return;
      }
      this.height = this.$refs.jsform.clientHeight + "px";
      this.loading = true;
      console.vlog("Deleting datapoint at ", this.datapoint.t);
      let res = await this.$frontend.rest(
        "DELETE",
        `api/objects/${encodeURIComponent(this.object.id)}/timeseries`,
        { t: this.datapoint.t }
      );
      if (!res.response.ok) {
        console.error(res);
        this.alert = res.data.error_description;
        this.loading = false;
        return;
      }
      this.loading = false;
      this.$emit("input", false);
    },
  },
  watch: {
    timestamp(v) {
      this.timestamp_changes++;
    },
    newdata(v) {
      this.data_changes++;
    },
    duration(v) {
      this.duration_changes++;
    },
  },
  computed: {
    editDuration() {
      return this.showDuration || this.duration != 0;
    },
    defaultIcon() {
      return (
        this.$store.state.heedy.object_types["timeseries"].icon ||
        "brightness_1"
      );
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
};
</script>