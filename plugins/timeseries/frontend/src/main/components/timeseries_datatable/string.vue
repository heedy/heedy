<template>
  <v-tooltip bottom>
    <template v-slot:activator="{ on, attrs }">
      <div
        style="
          width: 100%;
          overflow-x: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          text-align: center;
        "
        v-bind="attrs"
        v-on="on"
      >
        <span>{{ value }}</span>
      </div>
    </template>
    <span>{{ value }}</span>
  </v-tooltip>
</template>
<script>
export default {
  props: {
    value: String,
    column: Object,
  },
};
</script>