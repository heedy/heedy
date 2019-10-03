<template>
  <h-page-container noflex>
    <v-flex
      justify-center
      align-center
      text-center
      style="padding: 10px; padding-bottom: 20px;padding-top:20px;"
    >
      <h1 style="color:#1976d2;">{{ title }}</h1>
    </v-flex>
    <v-flex>
      <v-card>
        <div style="padding: 10px; padding-bottom: 0;">
          <v-alert
            v-if="alert.length>0"
            text
            outlined
            color="deep-orange"
            icon="error_outline"
          >{{ alert }}</v-alert>
        </div>
        <slot></slot>
      </v-card>
    </v-flex>
  </h-page-container>
</template>
<script>
export default {
  props: {
    title: String,
    alert: {
      type: String,
      default: ""
    }
  }
};
</script>