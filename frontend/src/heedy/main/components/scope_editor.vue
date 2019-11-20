<template>
  <v-combobox
    @input="oninput"
    :value="model"
    :filter="filter"
    :hide-no-data="!search"
    :items="items"
    :search-input.sync="search"
    hide-selected
    label="Permissions"
    multiple
    small-chips
    outlined
  >
    <template v-slot:no-data>
      <v-list-item>
        <span class="subheading" style="margin-right: 10px;color: gray;">Add Custom Scope:</span>
        <v-chip :color="`gray lighten-3`" label small>{{ search }}</v-chip>
      </v-list-item>
    </template>
    <template v-slot:selection="{ attrs, item, parent, selected }">
      <v-tooltip top>
        <template v-slot:activator="{ on }">
          <v-chip
            v-if="item === Object(item)"
            v-bind="attrs"
            :color="`${item.color} lighten-3`"
            :input-value="selected"
            label
            small
            v-on="on"
          >
            <span class="pr-2">{{ item.text }}</span>
            <v-icon small @click="parent.selectItem(item)">close</v-icon>
          </v-chip>
        </template>
        <span>{{ item.description }}</span>
      </v-tooltip>
    </template>
    <template v-slot:item="{ index, item }">
      <v-chip :color="`${item.color} lighten-3`" dark label small>{{ item.text }}</v-chip>
      <span style="font-size: 70%; color: gray; margin-left: 10px;">{{item.description }}</span>
    </template>
  </v-combobox>
</template>
<script>
export default {
  props: {
    value: String
  },
  data: () => ({
    search: "",
    customItems: [],
    loading: true
  }),
  created() {
    if (this.$store.state.heedy.appScopes == null) {
      this.$store.dispatch("getAppScopes");
    }
  },
  computed: {
    items() {
      let cscopes = this.$store.state.heedy.appScopes;
      if (cscopes == null) {
        return [{ header: "Loading..." }];
      }
      if (this.loading) {
        // If it was loading earlier, it created custom items. We want to clear those now
        this.customItems = [];
        this.loading = false;
      }

      let recognizedSet = {};
      let res = [];
      // We can make the permissions look pretty by splitting them into subtypes
      [
        {
          name: "self",
          description: "The app's permissions for its own private data",
          color: "green"
        },
        {
          name: "objects",
          description: "The app's access to objects belonging to you",
          color: "blue"
        },
        {
          name: "shared",
          description: "The app's access to objects shared with you",
          color: "purple"
        },
        {
          name: "owner",
          description: "The app's access to you",
          color: "orange"
        },
        {
          name: "users",
          description: "The app's access to other users",
          color: "red"
        }
      ].forEach(t => {
        let tScopes = Object.keys(cscopes).filter(v => v.startsWith(t.name));
        tScopes.map(k => {
          recognizedSet[k] = true;
        });
        res.push({ header: t.description, text: tScopes.join(" ") });
        res = res.concat(
          tScopes.map(k => ({
            text: k,
            description: cscopes[k],
            color: t.color
          }))
        );
      });

      let unrecognizedDB = Object.keys(cscopes)
        .filter(v => !(v in recognizedSet))
        .map(k => ({ text: k, description: cscopes[k], color: "gray" }));

      // Now add any unknown scopes
      if (this.customItems.length > 0 || unrecognizedDB.length > 0) {
        unrecognizedDB = unrecognizedDB.concat(this.customItems);
        res.push({
          header: "Custom & Unrecognized Scopes:",
          text: unrecognizedDB.map(k => k.text).join(" ")
        });
        res = res.concat(unrecognizedDB);
      }

      return res;
    },
    model() {
      if (this.value.length == 0) return [];
      let elements = this.value.split(" ");

      return elements.map(v => {
        let existingValue = null;
        this.items.map(k => {
          if (k.text == v) {
            existingValue = k;
          }
        });
        if (existingValue != null) {
          return existingValue;
        }
        v = {
          text: v,
          description: "unrecognized scope",
          color: "gray"
        };
        this.customItems.push(v);
        return v;
      });
    }
  },
  methods: {
    filter(item, queryText, itemText) {
      const hasValue = val => (val != null ? val : "");

      const text = hasValue(itemText);
      const query = hasValue(queryText);

      return (
        text
          .toString()
          .toLowerCase()
          .indexOf(query.toString().toLowerCase()) > -1
      );
    },
    oninput(event) {
      this.search = "";
      this.$emit(
        "input",
        event
          .map(v => {
            if (typeof v === "string") {
              return v;
            }
            return v.text;
          })
          .join(" ")
      );
    }
  }
};
</script>