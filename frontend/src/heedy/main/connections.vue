<template>
  <h-page-container>
    <v-card>
      <v-card-title>
        <v-list-item two-line>
          <v-btn color="blue darken-2" dark fab right absolute to="/create/connection">
            <v-icon>add</v-icon>
          </v-btn>
          <v-list-item-content>
            <v-list-item-title class="headline mb-1">Connections</v-list-item-title>
            <v-list-item-subtitle>External services and apps connected to heedy</v-list-item-subtitle>
          </v-list-item-content>
        </v-list-item>
      </v-card-title>
      <v-container>
        <div v-if="loading" style="color: gray; text-align: center;">Loading...</div>
        <div
          v-else-if="connections.length==0"
          style="color: gray; text-align: center;"
        >You don't have any connections.</div>
        <v-row no-gutters v-else>
          <v-col v-for="c in connections" :key="c.id" xs="12" sm="12" md="6" lg="4" xl="4">
            <v-card class="pa-2" outlined tile>
              <v-list-item two-line subheader :to="`/connections/${c.id}`">
                <v-list-item-avatar>
                  <h-icon :image="c.icon" :colorHash="c.id"></h-icon>
                </v-list-item-avatar>
                <v-list-item-content>
                  <v-list-item-title>{{ c.name }}</v-list-item-title>
                  <v-list-item-subtitle>{{ c.description }}</v-list-item-subtitle>
                </v-list-item-content>
              </v-list-item>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-card>
  </h-page-container>
</template>
<script>
export default {
  computed: {
    loading() {
      return this.$store.state.heedy.connections == null;
    },
    connections() {
      let c = this.$store.state.heedy.connections;

      return Object.keys(c).map(k => c[k]);
    }
  },
  created() {
    this.$store.dispatch("listConnections");
  }
};
</script>