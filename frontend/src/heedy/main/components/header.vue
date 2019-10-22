<template>
  <v-flex>
    <v-card>
      <v-card-title>
        <v-list-item two-line style="margin-top: -10px;margin-bottom: -8px">
          <v-list-item-avatar>
            <h-icon :image="icon" :colorHash="colorHash" />
          </v-list-item-avatar>
          <v-list-item-content>
            <v-list-item-title class="headline mb-1">{{ name }}</v-list-item-title>
            <v-list-item-subtitle>{{ description }}</v-list-item-subtitle>
          </v-list-item-content>
          <v-list-item-action>
            <v-toolbar flat collapsed>
              <v-toolbar-items style="margin-right: -30px;">
                <slot></slot>
              </v-toolbar-items>
            </v-toolbar>
          </v-list-item-action>
        </v-list-item>
      </v-card-title>
    </v-card>
  </v-flex>
</template>
<script>
export default {
  props: {
    icon: String,
    colorHash: {
      type: String,
      default: ""
    },
    name: String,
    description: String
  }
};
</script>