<template>
  <div>
    <v-select
      :prepend-icon="icon ? 'event' : ''"
      :items="selectItems"
      v-model="selectValue"
      class="fit"
    ></v-select>
    <v-dialog v-if="dialog" v-model="dialog" max-width="600">
      <v-card>
        <v-card-title
          class="headline grey lighten-2"
          primary-title
          style="font-size: 1em !important"
        >
          <v-icon left v-if="!$vuetify.breakpoint.xs">event</v-icon>
          {{ dialogTitle }}
        </v-card-title>
        <v-tabs v-model="dialogTab" grow>
          <v-tab> Time Range </v-tab>
          <v-tab> Relative </v-tab>
        </v-tabs>
        <v-tabs-items v-model="dialogTab">
          <v-tab-item>
            <vc-date-picker
              mode="dateTime"
              is-range
              is-expanded
              :toPage="calendarOptions.toPage"
              :attributes="calendarOptions.attributes"
              :columns="calendarOptions.columns"
              v-model="timeRange"
            />
          </v-tab-item>
          <v-tab-item>
            <v-card-text>
              <v-row v-if="allowIndex">
                <v-col cols="12" xs="6" sm="8" md="8">
                  <v-combobox
                    v-model="dialogRelativeNumber"
                    :items="dialogRelativeNumbers"
                  />
                </v-col>
                <v-col cols="12" xs="6" sm="4" md="4">
                  <v-select
                    :items="dialogRelativeItems"
                    v-model="dialogRelativeItem"
                  />
                </v-col>
              </v-row>
            </v-card-text>
          </v-tab-item>
        </v-tabs-items>

        <v-divider></v-divider>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="secondary" text @click="dialog = false">Cancel</v-btn>
          <v-btn
            color="primary"
            text
            @click="setDialogQuery"
            :disabled="!canSet"
            >Set</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>
<script>
import moment from "../../../dist/moment.mjs";

let fmtString = "YYYY-MM-DD HH:mm";
function parseTime(ts) {
  let tsf = parseFloat(ts);
  if (!isNaN(tsf)) {
    return moment.unix(tsf).format(fmtString);
  }
  if (ts == "now") {
    return "now";
  }
  if (ts.startsWith("now-")) {
    return `${ts.substring("now-".length, ts.length)} ago`;
  }
  return moment(ts).format(fmtString);
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
    if (q.i1 !== undefined && q.t1 === undefined) {
      let ival = parseInt(q.i1);
      if (ival < 0) {
        return `Last ${(-ival).toLocaleString()} datapoints${append}`;
      }
    }
    if (q.i1 === undefined && q.t1 !== undefined && q.t1.startsWith("now-")) {
      return `Last ${q.t1.substring("now-".length, q.t1.length)}${append}`;
    }
    if (q.i1 === undefined && q.t1 !== undefined) {
      return `${parseTime(q.t1)} ⇨ now`;
    }
  }
  if (
    q.t1 !== undefined &&
    q.t2 !== undefined &&
    q.i1 === undefined &&
    q.i2 === undefined
  ) {
    return `${parseTime(q.t1)} ⇨ ${parseTime(q.t2)}${append}`;
  }
  if (q.t1 === undefined && q.t2 === undefined) {
    if (q.i1 !== undefined && q.i2 !== undefined) {
      return `#${q.i1} ⇨ #${q.i2}${append}`;
    }
    if (q.i1 === undefined && q.i2 !== undefined) {
      return `First ${q.i2} datapoints${append}`;
    }
  }

  let tstring = "";
  if (q.i1 !== undefined && q.i1 != 0) {
    tstring = tstring + `i1:${q.i1};`;
  }
  if (q.i2 !== undefined) {
    tstring = tstring + `i2:${q.i2};`;
  }
  if (q.t1 !== undefined) {
    tstring = tstring + `t1:${parseTime(q.t1)};`;
  }
  if (q.t2 !== undefined) {
    tstring = tstring + `t2:${parseTime(q.t2)};`;
  }
  let val = tstring + append;
  if (val == "") {
    val = "All Time";
  }
  return val;
}

export default {
  props: {
    value: Object,
    allowIndex: {
      type: Boolean,
      default: true,
    },
    icon: {
      type: Boolean,
      default: true,
    },
  },
  data: () => ({
    dialogTab: 0,
    dialog: false,
    timeRange: { start: null, end: null },
    calendarOptions: null,
    // items using index queries
    indexItems: [{ i1: -1000 }, { i1: -100000 }],
    items: [
      {
        i1: 0,
      },
      {
        t1: "now-1d",
      },
      {
        t1: "now-1w",
      },
      {
        t1: "now-1mo",
      },
      {
        t1: "now-3mo",
      },
    ],
    dialogRelativeItems: [
      { text: "Datapoints", value: "dp" },
      { text: "Minutes", value: "m" },
      { text: "Hours", value: "h" },
      { text: "Days", value: "d" },
      { text: "Weeks", value: "w" },
      { text: "Months", value: "mo" },
      { text: "Years", value: "y" },
    ],
    dialogRelativeNumbers: [
      "1",
      "2",
      "3",
      "4",
      "6",
      "7",
      "10",
      "12",
      "15",
      "30",
      "45",
      "60",
      "90",
      "120",
      "1000",
      "10,000",
      "100,000",
      "1,000,000",
    ],
    dialogRelativeNumber: "1000",
    dialogRelativeItem: "dp",
  }),
  methods: {
    setDialogQuery() {
      let q = this.dialogQuery();
      if (q !== null) {
        this.dialog = false;
        this.setInput(q);
      }
    },
    setInput(q) {
      // We want to send the full query with just the relevant elements exchanged
      let newq = { ...this.value };

      // Remove all existing range info
      delete newq.t1;
      delete newq.t2;
      delete newq.i1;
      delete newq.i2;
      delete newq.t;
      delete newq.i;
      delete newq.limit;

      this.$emit("input", { ...newq, ...q });
    },
    showDialog() {
      let t = new Date();
      this.calendarOptions = {
        toPage: {
          month: t.getMonth() + 1,
          year: t.getFullYear(),
        },
        columns: this.$vuetify.breakpoint.xs ? 1 : 2,
        attributes: [
          {
            dot: true,
            dates: t,
          },
        ],
      };

      this.dialog = true;
    },
    dialogQuery() {
      if (
        this.dialogTab == 0 &&
        this.timeRange.start != null &&
        this.timeRange.end != null
      ) {
        return {
          t1: moment(this.timeRange.start).unix(),
          t2: moment(this.timeRange.end).unix(),
        };
      }
      if (this.dialogTab == 1) {
        let counter = parseInt(this.dialogRelativeNumber.replace(/,/g, ""), 10);
        if (!isNaN(counter) && counter > 0) {
          if (this.dialogRelativeItem == "dp") {
            return {
              i1: -counter,
            };
          }
          return {
            t1: `now-${counter}${this.dialogRelativeItem}`,
          };
        }
      }
      return null;
    },
  },
  computed: {
    dialogTitle() {
      let dq = this.dialogQuery();
      if (dq != null) {
        return queryLabel(dq);
      }
      return "Custom Range";
    },
    canSet() {
      return this.dialogQuery() != null;
    },
    selectItems() {
      let items = [
        ...this.items.map((q, i) => ({ text: queryLabel(q), value: i })),
        {
          text: "Custom",
          value: "custom",
        },
      ];
      if (this.allowIndex) {
        items = [
          ...this.indexItems.map((q, i) => ({
            text: queryLabel(q),
            value: i + this.items.length,
          })),
          ...items,
        ];
      }
      return items;
    },
    selectValue: {
      get() {
        if (this.dialog) {
          return "custom";
        }
        let ql = queryLabel(this.value);
        for (let i = 0; i < this.selectItems.length; i++) {
          if (this.selectItems[i].text == ql) {
            return this.selectItems[i].value;
          }
        }
        return "custom";
      },
      set(v) {
        if (v == "custom") {
          this.showDialog();
          return;
        }
        if (v < this.items.length) {
          this.setInput({
            ...this.items[v],
          });
          return;
        }
        this.setInput({
          ...this.indexItems[v - this.items.length],
        });
      },
    },
  },
  watch: {
    value(o) {
      let ql = queryLabel(o);
      for (let i = 0; i < this.selectItems.length; i++) {
        if (this.selectItems[i].text == ql) {
          return;
        }
      }
      if (o.i1 !== undefined || o.i2 !== undefined || o.i !== undefined) {
        this.indexItems.push(o);
      } else {
        this.items.push(o);
      }
    },
  },
  created() {
    let ql = queryLabel(this.value);
    for (let i = 0; i < this.selectItems.length; i++) {
      if (this.selectItems[i].text == ql) {
        return;
      }
    }
    this.items.push(this.value);
  },
};
</script>
<style>
.v-select.fit {
  font-size: 100%;
}
.v-select.fit .v-select__selections {
  margin-right: -30px;
}
.v-select.fit .v-select__selections input {
  width: 30px;
}
.v-select.fit .v-select__selection--comma {
  margin-right: 0;
}
</style>