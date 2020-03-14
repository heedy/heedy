<template>
  <div>
    <v-flex v-if="alert.length>0">
      <div style="padding: 10px; padding-bottom: 0;">
        <v-alert text outlined color="deep-orange" icon="error_outline">{{ alert }}</v-alert>
      </div>
    </v-flex>
    <v-flex>
      <v-data-table
        fixed-header
        :search="search"
        :headers="headers"
        :items="userItems"
        :loading="userItems.length==0"
        loading-text="Loading users..."
        disable-sort
      >
        <template v-slot:item.action="{ item }">
          <v-icon small class="mr-2" @click="editUser(item)">edit</v-icon>
          <v-icon small @click="delUser(item)">delete</v-icon>
        </template>
        <template v-slot:top>
          <v-toolbar flat color="white">
            <v-text-field
              v-model="search"
              append-icon="search"
              label="Search"
              single-line
              hide-details
            ></v-text-field>
            <div class="flex-grow-1"></div>
            <v-dialog v-model="createDialog" max-width="500px">
              <template v-slot:activator="{ on }">
                <v-btn color="primary" dark class="mb-2" v-on="on">Add User</v-btn>
              </template>
              <v-card>
                <v-card-title>
                  <span class="headline">Add User</span>
                </v-card-title>
                <v-card-text>
                  <v-container>
                    <v-row>
                      <v-col cols="12" sm="12" md="12">
                        <h3>Username</h3>
                      </v-col>
                      <v-col cols="12" sm="12" md="12">
                        <v-text-field label="Username" v-model="creating.username"></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="12" md="12">
                        <h3>Password</h3>
                      </v-col>
                      <v-col cols="12" sm="6" md="6">
                        <v-text-field type="password" v-model="creating.password" label="Password"></v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" md="6">
                        <v-text-field
                          type="password"
                          v-model="creating.password2"
                          label="Repeat Password"
                        ></v-text-field>
                      </v-col>
                    </v-row>
                  </v-container>
                </v-card-text>

                <v-card-actions>
                  <div class="flex-grow-1"></div>
                  <v-btn
                    color="blue darken-1"
                    text
                    @click="() => {creating={username: '',password: '',password2: ''};createDialog=false;}"
                  >Cancel</v-btn>
                  <v-btn color="blue darken-1" text @click="createUser()">Save</v-btn>
                </v-card-actions>
              </v-card>
            </v-dialog>
          </v-toolbar>
        </template>
      </v-data-table>
      <v-dialog v-model="updateDialog" max-width="500px">
        <v-card>
          <v-card-title>
            <span class="headline">Update {{ updating.id }}</span>
          </v-card-title>
          <v-card-text>
            <v-container>
              <v-row>
                <v-col cols="12" sm="12" md="12">
                  <h3>Username</h3>
                </v-col>
                <v-col cols="12" sm="12" md="12">
                  <v-text-field label="Username" v-model="updating.username"></v-text-field>
                </v-col>
                <v-col cols="12" sm="12" md="12">
                  <h3>Password</h3>
                </v-col>
                <v-col cols="12" sm="6" md="6">
                  <v-text-field type="password" label="Reset Password" v-model="updating.password"></v-text-field>
                </v-col>
                <v-col cols="12" sm="6" md="6">
                  <v-text-field
                    type="password"
                    label="Repeat Password"
                    v-model="updating.password2"
                  ></v-text-field>
                </v-col>
                <v-col cols="12" sm="12" md="12">
                  <h3>Admin</h3>
                </v-col>
                <v-col cols="12" sm="12" md="12">
                  <v-checkbox label="Admin" v-model="updating.admin"></v-checkbox>
                </v-col>
              </v-row>
            </v-container>
          </v-card-text>

          <v-card-actions>
            <div class="flex-grow-1"></div>
            <v-btn color="blue darken-1" text @click="updateDialog=false">Cancel</v-btn>
            <v-btn color="blue darken-1" text @click="updateUser()">Save</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-flex>
  </div>
</template>
<script>
export default {
  data: () => ({
    createDialog: false,
    updateDialog: false,
    creating: {
      username: "",
      password: "",
      password2: ""
    },
    updating: {
      id: "",
      password: "",
      password2: "",
      username: "",
      admin: false
    },
    search: "",
    admin: [],
    users: [],
    alert: "",
    headers: [
      { text: "Username", value: "username" },
      { text: "Name", value: "name" },
      { text: "Admin", value: "admin" },
      { text: "Actions", value: "action", align: "right", sortable: false }
    ]
  }),
  computed: {
    userItems() {
      let uv = this.users.map(u => ({
        username: u.username,
        name: u.name,
        admin: this.admin.includes(u.username) ? "admin" : ""
      }));
      return uv;
    }
  },
  methods: {
    editUser(u) {
      this.updating = {
        id: u.username,
        id_admin: u.admin == "admin",
        username: u.username,
        password: "",
        password1: "",
        admin: u.admin == "admin"
      };
      this.updateDialog = true;
    },
    createUser: async function() {
      let c = this.creating;
      this.creating = {
        username: "",
        password: "",
        password2: ""
      };
      this.createDialog = false;
      if (c.password != c.password2) {
        this.alert = "Passwords don't match";
        return;
      }
      let res = await this.$frontend.rest("POST", `/api/users`, {
        username: c.username,
        password: c.password
      });
      if (!res.response.ok) {
        this.alert = res.data.error_description;
      } else {
        this.alert = "";
      }
      this.reload();
    },
    delUser: async function(u) {
      if (
        confirm(
          `Are you sure you want to delete '${u.username}'? This action is irreversible, and all data associated with the account will be removed.`
        )
      ) {
        let res = await this.$frontend.rest(
          "DELETE",
          `/api/users/${u.username}`
        );
        if (!res.response.ok) {
          this.alert = res.data.error_description;
        } else {
          this.alert = "";
        }
        this.reload();
      }
    },
    updateUser: async function() {
      let toUpdate = {};
      if (this.updating.username != this.updating.id) {
        toUpdate.username = this.updating.username;
      }
      if (this.updating.password != "") {
        if (this.updating.password != this.updating.password2) {
          this.alert = "Passwords do not match";
          return;
        }
        toUpdate.password = this.updating.password;
      }
      this.updateDialog = false;
      if (Object.keys(toUpdate).length > 0) {
        let res = await this.$frontend.rest(
          "PATCH",
          `/api/users/${this.updating.id}`,
          toUpdate
        );
        if (!res.response.ok) {
          this.alert = res.data.error_description;
          return;
        }
      }
      if (this.updating.admin != this.updating.id_admin) {
        let res = await this.$frontend.rest(
          this.updating.admin ? "POST" : "DELETE",
          `/api/server/admin/${this.updating.username}`
        );
        if (!res.response.ok) {
          this.alert = res.data.error_description;
        } else {
          this.alert = "";
        }
      }
      this.reload();
    },
    reload: async function() {
      let u = this.$frontend.rest("GET", "/api/users").then(res => {
        if (!res.response.ok) {
          this.alert = res.data.error_description;
          this.users = [];
          return;
        }
        console.log("users", res.data);
        this.users = res.data;
      });
      let a = this.$frontend.rest("GET", "/api/server/admin").then(res => {
        if (!res.response.ok) {
          this.alert = res.data.error_description;
          this.admin = [];
          return;
        }
        console.log("admins", res.data);
        this.admin = res.data;
      });
    }
  },
  created() {
    this.reload();
  }
};
</script>