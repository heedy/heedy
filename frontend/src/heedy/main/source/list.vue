<template>
  <v-card>
    <v-container>
      <div v-if="sources.length==0" style="color: gray; text-align: center;">No Sources Found</div>
      <v-row no-gutters v-else>
        <v-col v-for="s in sources" :key="s.id" xs="12" sm="12" md="6" lg="4" xl="4">
          <v-card class="pa-2" outlined tile>
            <v-list-item two-line subheader :to="`/sources/${s.id}`">
              <v-list-item-avatar>
                <h-icon :image="s.icon" :colorHash="s.id"></h-icon>
              </v-list-item-avatar>
              <v-list-item-content>
                <v-list-item-title>{{ s.name }}</v-list-item-title>
                <v-list-item-subtitle>{{ s.description }}</v-list-item-subtitle>
              </v-list-item-content>
            </v-list-item>
          </v-card>
        </v-col>
      </v-row>
    </v-container>
  </v-card>
</template>
<script>
export default {
  props: {
    sources: Array
  }
};
</script>