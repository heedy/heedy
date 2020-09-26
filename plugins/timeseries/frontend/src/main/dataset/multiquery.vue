<template>
  <v-list flat style="background: none; padding-top: 0px" dense expand>
    <v-list-group
      v-for="(item, idx) in value"
      :key="idx"
      prepend-icon="timeline"
      value="true"
      no-action
    >
      <template v-slot:activator>
        <v-list-item-content>
          <v-list-item-title>Series {{ idx + 1 }}</v-list-item-title>
        </v-list-item-content>
      </template>
      <Query
        :value="item"
        :remove-series="value.length > 1"
        @input="(v) => setValue(idx, v)"
        @remove-series="() => removeSeries(idx)"
      ></Query>
    </v-list-group>
  </v-list>
</template>
<script>
import Query from "./query.vue";
export default {
  components: {
    Query,
  },
  props: {
    value: Array,
  },
  methods: {
    setValue(idx, v) {
      let vv = [...this.value];
      vv[idx] = v;
      this.$emit("input", vv);
    },
    removeSeries(idx) {
      let vv = [...this.value];
      vv.splice(idx, 1);
      this.$emit("input", vv);
    },
  },
};
</script>