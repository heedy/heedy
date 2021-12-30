<template>
  <div style="overflow-x: auto; whitespace: nowrap; border: 1px solid #ccc">
    <div
      :style="{
        width: '100%',
        minWidth: `${minWidth}px`,
      }"
    >
      <div
        :style="{
          borderBottom: 'solid 1px #ccc',
          display: 'grid',
          gridTemplateColumns: gridTemplateColumns,
          justifyContent: 'space-around',
          position: 'relative',
        }"
      >
        <v-tooltip bottom v-if="canEdit">
          <template v-slot:activator="{ on, attrs }">
            <div
              v-bind="attrs"
              v-on="on"
              style="
                position: absolute;
                top: 0;
                left: 0;
                height: 30px;
                width: 30px;
                text-align: center;
                padding-top: 2px;
              "
            >
              <v-icon x-small>edit</v-icon>
            </div>
          </template>
          <span
            >This timeseries is writable. Double-click on a row to edit
            data.</span
          >
        </v-tooltip>
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
        bench="1"
        :items="data"
        style="overflow-x: hidden"
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
            class="column-darken"
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
import getCol from "./datatable/columns.js";

import EditDialog from "./datatable/edit_dialog.vue";

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
      return getCol(col.type);
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
<style scoped>
.column-darken:hover {
  filter: brightness(93%);
}
</style>