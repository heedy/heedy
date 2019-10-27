<template>
  <h-header
    :icon="source.icon"
    :colorHash="source.id"
    :name="source.name"
    :description="source.description"
  >
    <v-select
      :items="queryOptions"
      v-model="query"
      style="padding-top: 17px;padding-right: 10px; max-width: 250px;"
      prepend-icon="event"
    ></v-select>
    <v-dialog v-model="dialog" max-width="500">
      <v-card>
        <v-card-title class="headline grey lighten-2" primary-title>Custom Query</v-card-title>

        <v-card-text>
          <v-row>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-text-field label="Start Index" />
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-text-field label="End Index" />
            </v-col>
          </v-row>
          <v-row>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-datetime-picker label="Start Time" />
            </v-col>
            <v-col cols="12" xs="12" sm="6" md="6">
              <v-datetime-picker label="End Time" />
            </v-col>
          </v-row>
          <v-row>
            <v-col cols="12" xs="12">
              <v-text-field outlined label="Transform" />
            </v-col>
          </v-row>
        </v-card-text>

        <v-divider></v-divider>

        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn color="secondary" text @click="dialog = false">Cancel</v-btn>
          <v-btn color="primary" text @click="dialog = false">Query</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-tooltip bottom>
      <template #activator="{on}">
        <v-btn icon v-on="on" @click="live = !live">
          <v-icon v-if="live" style="font-size: 1.25em;">fas fa-wifi</v-icon>
          <span v-else class="fa-stack fa-2x" style="font-size: 1.15em">
            <i class="fas fa-wifi fa-stack-1x"></i>
            <i class="fas fa-ban fa-stack-2x" style="color:Tomato"></i>
          </span>
        </v-btn>
      </template>
      <span>{{ live? "Disable live update": "Enable live update"}}</span>
    </v-tooltip>
    <v-tooltip bottom>
      <template #activator="{on}">
        <v-btn icon v-on="on" :to="`/sources/${source.id}/stream/update`">
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
export default {
  components: {
    VDatetimePicker
  },
  props: {
    source: Object
  },
  data: () => ({
    dialog: false,
    live: true,
    queryOptions: [
      { text: "Last 100 Datapoints", value: 0, q: { i1: -100 } },
      {
        text: "Last Week",
        value: 1,
        q: {
          t1: -1000
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
        if (this.$route.query.q === undefined) {
          return 0;
        }
        return this.$route.query.q;
      },
      set(v) {
        if (v == "custom") {
          console.log("CUSTOM!");
        }
        this.$router.replace({ query: this.queryOptions[v].q });
      }
    }
  },

  created() {}
};
</script>