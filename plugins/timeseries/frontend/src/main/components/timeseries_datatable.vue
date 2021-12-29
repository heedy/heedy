<template>
  <div style="overflow-x: scroll; whitespace: nowrap; border: 1px solid #ccc">
    <div :style="{ width: '100%', minWidth: `${minWidth}px` }">
      <div
        :style="{
          borderBottom: 'solid 1px #ccc',
          display: 'grid',
          gridTemplateColumns: gridTemplateColumns,
          justifyContent: 'space-around',
        }"
      >
        <div
          style="
            height: 30px;
            font-weight: bold;
            display: flex;
            justify-content: center;
            align-items: center;
          "
          v-for="col in columns"
          :key="col.prop"
        >
          {{ col.name }}
        </div>
      </div>
      <v-virtual-scroll
        :height="tblheight"
        item-height="30"
        :items="data"
        style="overflow-x: visible"
      >
        <template v-slot:default="{ index, item }">
          <div
            :style="{
              height: `30px`,
              display: 'grid',
              gridTemplateColumns: gridTemplateColumns,
              justifyContent: 'space-around',
              background: index % 2 === 0 ? '#f6f6f6' : 'white',
            }"
            @dblclick.prevent="() => editData(index)"
          >
            <div
              :style="{
                //border: 'solid 1px',
                height: '30px',
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
              }"
              v-for="col in columns"
              :key="col.prop"
            >
              <component
                :is="getColumn(col).component"
                :value="getValue(col, item)"
                :column="col"
              />
            </div>
          </div>
        </template>
      </v-virtual-scroll>
    </div>
    <edit-dialog
      v-if="editing"
      @input="cancelEdit"
      :object="object"
      :datapoint="datapoint"
    />
  </div>
</template>
<script>
import JSONCell from "./timeseries_datatable/json.vue";
import TimestampCell from "./timeseries_datatable/timestamp.vue";
import NumberCell from "./timeseries_datatable/number.vue";
import StringCell from "./timeseries_datatable/string.vue";
import BooleanCell from "./timeseries_datatable/boolean.vue";
import DurationCell from "./timeseries_datatable/duration.vue";

import EditDialog from "./timeseries_datatable/edit_dialog.vue";

const columnTypes = {
  timestamp: {
    component: TimestampCell,
    width: 180,
  },
  duration: {
    component: DurationCell,
    width: 80,
  },
  number: {
    component: NumberCell,
    width: 100,
  },
  string: {
    component: StringCell,
    width: 150,
  },
  boolean: {
    component: BooleanCell,
    width: 8 * 5,
  },
  enum: {
    component: StringCell,
    width: 100,
  },
  json: {
    component: JSONCell,
    width: 250,
  },
};

export default {
  components: {
    EditDialog,
  },
  props: {
    data: {
      type: Array,
      required: true,
    },
    columns: {
      type: Array,
      required: true,
    },
    timeseries: {
      type: String,
      default: "",
    },
    editable: {
      type: Boolean,
      default: false,
    },
  },
  data: () => ({
    editing: false,
    datapoint: null,
  }),
  methods: {
    getValue(col, item) {
      let s = col.prop.split(".");
      let res = item;
      for (let i = 0; i < s.length; i++) {
        res = res[s[i]];
        if (res === undefined) {
          return undefined;
        }
      }
      return res;
    },
    getColumn(col) {
      return columnTypes[col.type] || columnTypes.json;
    },
    editData(idx) {
      if (!this.canEdit) {
        return;
      }
      console.vlog("Editing Datapoint", idx, this.data[idx]);
      this.datapoint = this.data[idx];
      this.editing = true;
    },
    cancelEdit() {
      this.editing = false;
    },
  },
  computed: {
    tblheight() {
      return Math.min(this.data.length * 30, 350);
    },
    widthArray() {
      return this.columns.map((col) => {
        col = {
          ...this.getColumn(col),
          ...col,
        };
        let width = col.name.length * 8;
        if (col.width && col.width > width) {
          width = col.width;
        }
        return width;
      });
    },
    minWidth() {
      return this.widthArray.reduce((a, b) => a + b, 0);
    },
    gridTemplateColumns() {
      // We figure out the width to give each column, based on the settings on each column
      // type as well as heading size
      return this.widthArray.map((w) => `${w}px`).join(" ");
    },
    object() {
      if (this.timeseries == "" || this.timeseries == null) {
        return null;
      }
      return this.$store.state.heedy.objects[this.timeseries] || null;
    },
    canEdit() {
      if (!this.editable || this.object === null) {
        return false;
      }
      let access = this.object.access.split(" ");
      return (
        this.object.meta.schema.type !== undefined &&
        (access.includes("*") || access.includes("write"))
      );
    },
  },
  watch: {
    timeseries(newValue) {
      if (this.timeseries != "" && this.timeseries != null) {
        this.$store.dispatch("readObject", {
          id: newValue,
        });
      }
    },
  },
  created() {
    if (this.timeseries != "" && this.timeseries != null) {
      this.$store.dispatch("readObject", {
        id: this.timeseries,
      });
    }
  },
};
</script>