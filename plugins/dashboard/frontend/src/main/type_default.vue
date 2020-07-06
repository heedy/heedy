<template>
  <v-flex>
    <v-row>
      <v-col style="width:100%;text-align:center;">
        <h2 style="color: #c9c9c9;margin-top: 5%;">Unrecognized element type "{{element.type}}"</h2>
      </v-col>
    </v-row>
  </v-flex>
</template>
<script>
export default {
  props: {
    object: Object,
    element: Object
  }
};
</script>