<template>
  <h-dataset-visualization :query="query">
    <v-col v-if="haswrite" cols="12" sm="12" md="6" lg="6" xl="4">
      <v-card>
        <v-card-title>Insert</v-card-title>
        <v-card-text>
          <insert :object="object"></insert>
        </v-card-text>
      </v-card>
    </v-col>
  </h-dataset-visualization>
</template>
<script>
import Insert from "./insert.vue";
export default {
  components: {
    Insert,
  },
  props: {
    object: Object,
  },
  data: () => ({
    defaultQuery: { i1: -1000 },
    query: [],
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
  watch: {
    "$route.query": function(n, o) {
      if (Object.keys(n).length == 0) {
        this.$router.replace({ query: this.defaultQuery });
      } else {
        this.query = [
          {
            ...n,
            timeseries: this.object.id,
          },
        ];
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
      this.query = [
        {
          ...this.$route.query,
          timeseries: this.object.id,
        },
      ];
    }
  },
};
</script>
