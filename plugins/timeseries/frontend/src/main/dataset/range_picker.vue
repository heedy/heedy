<template>
  <div>
    <v-select
      :prepend-icon="icon ? 'event' : ''"
      :items="selectItems"
      v-model="selectValue"
    ></v-select>
    <v-dialog v-model="dialog" max-width="500">
      <v-card>
        <v-card-title class="headline grey lighten-2" primary-title
          >Custom Query</v-card-title
        >

        <v-card-text>
          <v-row v-if="allowIndex">
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-text-field label="Start Index" v-model="custom.i1" />
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-text-field label="End Index" v-model="custom.i2" />
            </v-col>
          </v-row>
          <v-row>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-text-field label="Start Time" v-model="custom.t1" />
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-text-field label="End Time" v-model="custom.t2" />
            </v-col>
          </v-row>
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="secondary" text @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" text @click="customquery">Set</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>
<script>
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
    dialog: false,
    custom: { i1: "", i2: "", t1: "", t2: "" },
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
  }),
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
        q.i1 = parseInt(this.custom.i1);
      }
      if (this.custom.i2 != "") {
        q.i2 = parseInt(this.custom.i2);
      }

      this.dialog = false;
      this.custom = { i1: "", i2: "", t1: "", t2: "" };
      this.setInput(q);
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
  },
  computed: {
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
          this.dialog = true;
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
      this.items.push(o);
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