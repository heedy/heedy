<template>
  <v-tooltip bottom>
    <template v-slot:activator="{ on, attrs }">
      <div
        :style="{
          overflowX: 'hidden',
          textOverflow: 'ellipsis',
          whiteSpace: 'nowrap',
          textAlign: align,
        }"
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
    align: {
      type: String,
      default: "center",
    },
  },
};
</script>