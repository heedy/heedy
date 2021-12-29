<template>
  <h-dataset-visualization :query="query">
    <v-col v-if="haswrite" cols="12" sm="12" md="6" lg="6" xl="4">
      <v-card>
        <v-card-title
          >Insert
          <v-spacer />
          <v-tooltip bottom>
            <template #activator="{ on }">
              <v-btn icon @click="toggleCustomTimestamp" v-on="on">
                <v-icon :color="customTimestamp ? 'primary' : undefined"
                  >more_time</v-icon
                >
              </v-btn>
            </template>
            <span
              >{{ customTimestamp ? "Remove " : "Use " }}Custom Timestamp</span
            >
          </v-tooltip>
        </v-card-title>
        <v-card-text>
          <h-timeseries-datapoint-inserter
            :object="object"
            :customTimestamp="customTimestamp"
            @inserted="resetCustomTimestamp"
          ></h-timeseries-datapoint-inserter>
        </v-card-text>
      </v-card>
    </v-col>
  </h-dataset-visualization>
</template>
<script>
function getQ(name, q) {
  let res = {};
  res[name] = q;
  return res;
}
export default {
  props: {
    object: Object,
  },
  data: () => ({
    defaultQuery: { i1: -1000 },
    query: {},
    customTimestamp: false,
  }),
  computed: {
    haswrite() {
      let access = this.object.access.split(" ");
      return (
        this.object.meta.schema.type !== undefined &&
        (access.includes("*") || access.includes("write"))
      );
    },
  },
  methods: {
    toggleCustomTimestamp() {
      this.customTimestamp = !this.customTimestamp;
    },
    resetCustomTimestamp() {
      this.customTimestamp = false;
    },
  },
  watch: {
    "$route.query": function (n, o) {
      if (Object.keys(n).length == 0) {
        this.$router.replace({ query: this.defaultQuery });
      } else {
        this.query = getQ(this.object.name, {
          ...n,
          timeseries: this.object.id,
        });
      }
    },
    object(n, o) {
      if (n.id != o.id) {
        this.query = {
          ...this.query,
          timeseries: n.id,
        };
      }
    },
  },
  created() {
    // If no query, use default
    if (Object.keys(this.$route.query).length == 0) {
      this.$router.replace({ query: this.defaultQuery });
    } else {
      this.query = getQ(this.object.name, {
        ...this.$route.query,
        timeseries: this.object.id,
      });
    }
  },
};
</script>
