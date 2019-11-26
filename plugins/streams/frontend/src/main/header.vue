<template>
  <h-header
    :icon="object.icon"
    :colorHash="object.id"
    :name="object.name"
    :description="object.description"
    :showTitle="!$vuetify.breakpoint.xs"
  >
    <v-select
      :items="queryOptions"
      v-model="query"
      style="padding-top: 17px;padding-right: 10px; max-width: 250px;"
      :prepend-icon="$vuetify.breakpoint.xs?'':'event'"
    ></v-select>
    <v-dialog v-model="dialog" max-width="500">
      <v-card>
        <v-card-title class="headline grey lighten-2" primary-title>Custom Query</v-card-title>

        <v-card-text>
          <v-row>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-text-field
                label="Start Index"
                :placeholder="this.$route.query.i1 || ''"
                v-model="custom.i1"
              />
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-text-field
                label="End Index"
                :placeholder="this.$route.query.i2 || ''"
                v-model="custom.i2"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-text-field
                label="Start Time"
                :placeholder="this.$route.query.t1 || ''"
                v-model="custom.t1"
              />
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-text-field
                label="End Time"
                :placeholder="this.$route.query.t2 || ''"
                v-model="custom.t2"
              />
            </v-col>
          </v-row>
          <!--
          <v-row>
            <v-col cols="12" xs="12">
              <v-text-field outlined label="Transform" />
            </v-col>
          </v-row>
          -->
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="secondary" text @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" text @click="customquery">Query</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-tooltip bottom>
      <template #activator="{on}">
        <v-btn icon v-on="on" :to="`/objects/${object.id}/stream/update`">
          <v-icon>edit</v-icon>
        </v-btn>
      </template>
      <span>Edit Stream</span>
    </v-tooltip>
  </h-header>
</template>
<script>
import moment from "../../dist/moment.mjs";
import VDatetimePicker from "vuetify-datetime-picker/src/components/DatetimePicker.vue";

function parseTime(ts) {
  let tsf = parseFloat(ts);
  if (!isNaN(tsf)) {
    return moment.unix(tsf).calendar();
  }
  if (ts == "now") {
    return "now";
  }
  if (ts.startsWith("now-")) {
    return `${ts.substring("now-".length, ts.length)} ago`;
  }
  return moment(ts).calendar();
}
function queryLabel(q) {
  let append = "";
  if (q.limit !== undefined) {
    append = ` (limit ${q.limit})`;
  }
  if (q.i !== undefined) {
    return `@${q.i}${append}`;
  }
  if (q.t !== undefined) {
    return `@${parseTime(q.t)}${append}`;
  }
  // If it is a starting query
  if (q.i2 === undefined && q.t2 === undefined) {
    if (q.i1 !== undefined && q.t1 === undefined && q.i1.startsWith("-")) {
      return `Last ${q.i1.substring(1, q.i1.length)} datapoints${append}`;
    }
    if (q.i1 === undefined && q.t1 !== undefined && q.t1.startsWith("now-")) {
      return `Last ${q.t1.substring("now-".length, q.t1.length)}${append}`;
    }
    if (q.i1 === undefined && q.t1 !== undefined) {
      return `${parseTime(q.t1)} - now`;
    }
  }
  if (
    q.t1 !== undefined &&
    q.t2 !== undefined &&
    q.i1 === undefined &&
    q.i2 === undefined
  ) {
    return `${parseTime(q.t1)} - ${parseTime(q.t2)}${append}`;
  }
  if (q.t1 === undefined && q.t2 === undefined) {
    if (q.i1 !== undefined && q.i2 !== undefined) {
      return `#${q.i1} - #${q.i2}${append}`;
    }
    if (q.i1 === undefined && q.i2 !== undefined) {
      return `First ${q.i2} datapoints${append}`;
    }
  }

  let tstring = "";
  if (q.i1 !== undefined) {
    tstring = tstring + `i1:${q.i1};`;
  }
  if (q.i2 !== undefined) {
    tstring = tstring + `i2:${q.i1};`;
  }
  if (q.t1 !== undefined) {
    tstring = tstring + `t1:${parseTime(q.t1)};`;
  }
  if (q.t2 !== undefined) {
    tstring = tstring + `t2:${parseTime(q.t2)};`;
  }
  return tstring + append;
}

export default {
  components: {
    VDatetimePicker
  },
  props: {
    object: Object
  },
  data: () => ({
    dialog: false,
    custom: { i1: "", i2: "", t1: "", t2: "" },
    live: true,
    queryOptions: [
      { text: "Last 100 datapoints", value: 0, q: { i1: "-100" } },
      { text: "Last 1000 datapoints", value: 1, q: { i1: "-1000" } },
      {
        text: "Last 1d",
        value: 2,
        q: {
          t1: "now-1d"
        }
      },
      {
        text: "Last 1w",
        value: 3,
        q: {
          t1: "now-1w"
        }
      },
      {
        text: "Last 1mo",
        value: 4,
        q: {
          t1: "now-1mo"
        }
      },
      {
        text: "Last 3mo",
        value: 5,
        q: {
          t1: "now-3mo"
        }
      },
      {
        text: "Custom",
        value: "custom"
      }
    ]
  }),
  computed: {
    query: {
      get() {
        console.log(this.$route.query);
        if (Object.keys(this.$route.query).length == 0) {
          this.$router.replace({ query: this.queryOptions[0].q });
          return 0;
        }

        let lbl = queryLabel(this.$route.query);
        for (let i = 0; i < this.queryOptions.length; i++) {
          if (this.queryOptions[i].text == lbl) {
            return i;
          }
        }
        // The given label doesn't exist, so add the query to the list
        this.queryOptions.splice(this.queryOptions.length - 1, 0, {
          text: lbl,
          value: this.queryOptions.length - 1,
          q: this.$route.query
        });
        return this.queryOptions.length - 2;
      },
      set(v) {
        if (v == "custom") {
          this.dialog = true;
          return;
        }
        console.log("SET", v);
        let lbl = queryLabel(this.$route.query);
        if (lbl != this.queryOptions[v].text) {
          this.$router.replace({ query: this.queryOptions[v].q });
        }
      }
    }
  },
  methods: {
    customquery() {
      let q = {};
      if (this.custom.t1 != "") {
        q.t1 = this.custom.t1;
      }
      if (this.custom.t2 != "") {
        q.t2 = this.custom.t2;
      }
      if (this.custom.i1 != "") {
        q.i1 = this.custom.i1;
      }
      if (this.custom.i2 != "") {
        q.i2 = this.custom.i2;
      }

      this.dialog = false;
      this.custom = { i1: "", i2: "", t1: "", t2: "" };

      this.$router.replace({ query: q });
    }
  }
};
</script>