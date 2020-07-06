<template>
  <v-flex style="padding-top: 0px;">
    <v-row>
      <v-col v-if="datavis.length==0" style="width: 100%; text-align: center;">
        <h1 style="color: #c9c9c9;margin-top: 5%;">{{ message }}</h1>
      </v-col>
      <v-col v-for="d in datavis" :key="d.id" cols="12" sm="12" md="6" lg="6" xl="4">
        <v-card>
          <v-card-title v-if="d.title !== undefined">{{ d.title }}</v-card-title>
          <v-card-text>
            <component :is="view(d.type)" :object="object" :element="d" />
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-flex>
</template>
<script>
import DefaultType from "./type_default.vue";

function reindex(arr) {
  for (let i = 0; i < arr.length; i++) {
    if (arr[i].index != i) {
      arr[i].index = i;
    }
  }
  return arr;
}

export default {
  props: {
    object: Object
  },
  data: () => ({
    message: "Loading...",
    datavis: [],
    subscribed: false
  }),
  methods: {
    view(t) {
      let vs = this.$store.state.dashboard.types;
      if (vs[t] === undefined) {
        return DefaultType;
      }
      return vs[t];
    },
    subscribe(id) {
      if (this.subscribed) {
        this.$frontend.dashboard.unsubscribe(id, "mainviews");
      }
      this.subscribed = true;
      this.message = "Loading...";
      this.datavis = [];
      this.$frontend.dashboard.subscribe(id, "mainviews", dv => {
        if (dv.query_status !== undefined) {
          // Special-case query status messages
          this.message = dv.query_status.data;
          return;
        }
        // Now perform the updates on datavis
        this.message = "Dashboard Empty";
        if (dv.length > 1) {
          this.datavis = dv; // If it has more than one element, it is just reading, so skip all checks
        } else {
          dv = dv[0];
          if (dv.delete !== undefined) {
            this.datavis = reindex(
              this.datavis.filter(el => el.id != dv.delete)
            );
            return;
          }
          if (dv.index >= this.datavis.length - 1) {
            let fil = this.datavis.filter(el => el.id != dv.id);
            fil.push(dv);
            this.datavis = reindex(fil);
            return;
          }
          if (this.datavis[dv.index].id != dv.id) {
            let fil = this.datavis.filter(el => el.id != dv.id);
            fil.splice(dv.index, 0, dv);
            this.datavis = reindex(fil);
            return;
          }

          // vue 2 derping
          let fil = this.datavis;
          fil[dv.index] = dv;
          this.datavis = [...fil];
          return;
        }
      });
    }
  },
  watch: {
    object(n, o) {
      if (n.id != o.id) {
        if (this.subscribed) {
          this.$frontend.dashboard.unsubscribe(o.id, "mainviews");
          this.subscribed = false;
          this.subscribe(n.id);
        }
      }
    }
  },
  created() {
    this.subscribe(this.object.id);
  },
  beforeDestroy() {
    this.$frontend.dashboard.unsubscribe(this.object.id, "mainviews");
  }
};
</script>