<template>
  <v-combobox
    :value="value"
    @input="(v) => $emit('input',(v==null?'':v))"
    outlined
    :label="'Choose ' + objname"
    :items="objects"
    :filter="filter"
    :search-input.sync="search"
    :hide-no-data="!search"
    :hide-details="true"
    :multiple="multiple"
    :hide-selected="true"
  >
    <template v-slot:no-data>
      <v-list-item dense two-line @click="() => addItem(search)">
        <v-list-item-avatar>
          <h-icon image="fas fa-question" colorHash="unknown" />
        </v-list-item-avatar>
        <span class="subheading" style="margin-right:5px;">{{ objname }} ID:</span>
        <v-chip label>{{ search }}</v-chip>
      </v-list-item>
    </template>
    <template v-slot:item="{on,item}">
      <v-list-item dense two-line v-on="on">
        <v-list-item-avatar>
          <h-icon
            :image="obj(item).icon"
            :defaultIcon="defaultIcon(obj(item).type)"
            :colorHash="item"
          />
        </v-list-item-avatar>
        <v-list-item-content>
          <v-list-item-title>{{obj(item).name}}</v-list-item-title>
          <v-list-item-subtitle>
            <v-chip-group v-if="obj(item).tags!=''">
              <v-chip v-for="tag in obj(item).tags.split(' ')" :key="tag" x-small disabled>{{ tag }}</v-chip>
            </v-chip-group>
            <div v-else>{{ item.description }}</div>
          </v-list-item-subtitle>
        </v-list-item-content>
        <v-list-item-action></v-list-item-action>
      </v-list-item>
    </template>
    <template v-slot:selection="{item,attrs}">
      <v-chip :key="item" v-bind="attrs" label @click="()=>chipClick(item)">
        <h-icon
          :size="25"
          :image="obj(item).icon"
          :defaultIcon="defaultIcon(obj(item).type)"
          :colorHash="item"
          style="margin-right: 10px;"
        />
        {{ obj(item).name }}
      </v-chip>
    </template>
  </v-combobox>
</template>
<script>
export default {
  props: {
    owner: {
      type: String,
      default: ""
    },
    type: {
      type: String,
      default: ""
    },
    app: {
      type: String,
      default: ""
    },
    access: {
      type: String,
      default: ""
    },
    multiple: {
      type: Boolean,
      default: false
    },
    value: [String, Array]
  },
  data: () => ({
    search: "",
    extra: []
  }),
  computed: {
    objname() {
      if (
        this.type != "" &&
        this.$store.state.heedy.object_types[this.type] !== undefined
      ) {
        return (
          this.$store.state.heedy.object_types[this.type].title || "Object"
        );
      }
      return "Object";
    },
    objects() {
      let objs = [];
      if (this.owner == "") {
        objs = Object.values(this.$store.state.heedy.objects);
      } else {
        objs = Object.keys(
          this.$store.state.heedy.userObjects[this.owner] || {}
        ).map(id => this.$store.state.heedy.objects[id]);
      }
      objs = objs.filter(o => o != null);

      if (this.app != "") {
        objs = objs.filter(o => o.app == this.app);
      }

      if (this.type != "") {
        objs = objs.filter(o => o.type == this.type);
      }
      if (this.access != "") {
        let tac = this.access.split(" ");
        objs = objs.filter(o => {
          let oa = o.access.split(" ");
          if (oa.includes("*")) {
            return true;
          }
          if (tac.includes("*")) {
            return false;
          }
          return tac.every(scope => oa.includes(scope));
        });
      }

      objs = objs.map(obj => obj.id);
      if (this.extra.length > 0) {
        this.extra = this.extra.filter(e => !objs.includes(e));
        objs = [...objs, ...this.extra];
      }
      return objs;
    }
  },
  methods: {
    addItem(id) {
      this.search = "";
      if (!this.multiple) {
        this.$emit("input", id);
        return;
      }
      this.$emit("input", [...this.value, id]);
    },
    chipClick(id) {
      this.search = "";
      if (!this.multiple) {
        this.$emit("input", "");
        return;
      }
      this.$emit(
        "input",
        this.value.filter(e => e != id)
      );
    },
    filter(iid, queryText, itemText) {
      let item = this.obj(iid);

      let queryWords = queryText.toLowerCase().split(" ");
      let tags = item.tags.toLowerCase().split(" ");
      let name = item.name.toLowerCase();

      return queryWords.every(
        w =>
          name.includes(w) ||
          item.id.includes(w) ||
          !tags.every(t => !t.includes(w)) ||
          (item.type.includes(w) && this.type == "")
      );
    },
    obj(id) {
      if (
        this.$store.state.heedy.objects[id] === undefined ||
        this.$store.state.heedy.objects[id] == null
      ) {
        return {
          id: id,
          name: this.objname + " ID: " + id,
          icon: "fas fa-question",
          type: "",
          app: null,
          description: "",
          tags: ""
        };
      }
      return this.$store.state.heedy.objects[id];
    },
    defaultIcon(otype) {
      let oti = this.$store.state.heedy.object_types[otype] || {
        icon: "assignment"
      };
      return oti.icon;
    }
  },
  watch: {
    owner: function(o) {
      if (o != "") {
        this.$store.dispatch("readUserObjects", { username: o });
      } else if (this.$store.state.app.info.user != null) {
        this.$store.dispatch("readUserObjects", {
          username: this.$store.state.app.info.user.username
        });
      }
    },
    value(olist) {
      if (olist == null || olist == "") {
        return;
      }
      if (typeof olist === "string") {
        olist = [olist];
      }
      olist.forEach(oid => {
        if (oid != "" && oid != null) {
          if (
            this.$store.state.heedy.objects[oid] === undefined ||
            this.$store.state.heedy.objects[oid] == null
          ) {
            // The object doesn't exist in the cache. Add it to the extra, and query it
            this.$store.dispatch("readObject", {
              id: oid
            });
            if (!this.extra.includes(oid)) {
              this.extra.push(oid);
            }
          }
        }
      });
    }
  },

  created() {
    if (this.owner != "") {
      this.$store.dispatch("readUserObjects", { username: this.owner });
    } else if (this.$store.state.app.info.user != null) {
      this.$store.dispatch("readUserObjects", {
        username: this.$store.state.app.info.user.username
      });
    }
  }
};
</script>