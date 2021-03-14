<template>
  <div>
    <v-flex v-if="alert.length > 0">
      <div style="padding: 10px; padding-bottom: 0">
        <v-alert text outlined color="deep-orange" icon="error_outline">{{
          alert
        }}</v-alert>
      </div>
    </v-flex>
    <v-flex>
      <v-data-table
        fixed-header
        :headers="headers"
        :items="sessions"
        :loading="sessions.length == 0"
        loading-text="Loading sessions..."
      >
        <template v-slot:item.action="{ item }">
          <v-icon small @click="delSession(item)">delete</v-icon>
        </template>
      </v-data-table>
    </v-flex>
  </div>
</template>
<script>
export default {
  data: () => ({
    alert: "",
    sessions: [],
    firstReload: true,
    headers: [
      { text: "Description", value: "description" },
      { text: "Created", value: "created_date" },
      { text: "Last Used", value: "last_access_date" },
      { text: "Actions", value: "action", align: "right", sortable: false },
    ],
  }),
  methods: {
    delSession: async function (s) {
      if (confirm(`Are you sure you want to log out the given session?`)) {
        let uname = this.$store.state.app.info.user.username;
        let res = await this.$frontend.rest(
          "DELETE",
          `/api/users/${uname}/sessions/${s.sessionid}`
        );
        if (!res.response.ok) {
          this.alert = res.data.error_description;
          return;
        }
        this.alert = "";
        this.reload();
      }
    },
    reload: async function () {
      let uname = this.$store.state.app.info.user.username;
      let res = await this.$frontend.rest(
        "GET",
        `/api/users/${uname}/sessions`
      );
      if (!res.response.ok) {
        this.alert = res.data.error_description;
        this.users = [];
        if (!this.firstReload) {
          location.reload(true); // Reload the
        }
        this.firstReload = false;
        return;
      }
      this.alert = "";
      this.sessions = res.data;
      this.firstReload = false;
    },
  },
  created() {
    this.reload();
  },
};
</script>
