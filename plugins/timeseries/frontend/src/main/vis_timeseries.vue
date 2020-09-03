<template>
  <h-dataset-visualization :query="query"></h-dataset-visualization>
</template>
<script>
export default {
  props: {
    object: Object,
  },
  data: () => ({
    defaultQuery: { i1: -1000 },
    query: [],
  }),
  watch: {
    "$route.query": function (n, o) {
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
