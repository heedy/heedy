<template>
  <v-card color="rgb(245,245,250)" elevation="24">
    <v-card-title text-center>
      <v-layout row justify-center>
        <v-flex text-center style="padding-top: 1cm;">
          <h1 style="color:#1976d2;padding-bottom: 7px;">heedy</h1>
          <h4>Create your Database</h4>
        </v-flex>
      </v-layout>
    </v-card-title>
    <v-form @submit="submit" style="padding-left:15px;padding-right:15px;">
      <v-card-text>
        <v-alert
          v-if="alert.length>0"
          text
          outlined
          color="deep-orange"
          icon="error_outline"
        >{{ alert }}</v-alert>
        <v-alert v-if="success.length>0" text outlined type="success">{{ success }}</v-alert>
        <v-container>
          <v-layout row wrap>
            <v-flex xs12>
              <h3>Username</h3>
              <v-text-field
                label="Username"
                placeholder="admin"
                v-model.trim="username"
                required
                autofocus
                solo
                tabindex="1"
              ></v-text-field>
            </v-flex>
            <v-flex xs12>
              <h3>Password</h3>
            </v-flex>
            <v-flex md6 xs12>
              <v-text-field
                label="Password"
                placeholder="Password"
                type="password"
                v-model="password1"
                required
                solo
                tabindex="2"
              ></v-text-field>
            </v-flex>
            <v-flex md6 xs12>
              <v-text-field
                label="Repeat Password"
                placeholder="Repeat Password"
                type="password"
                v-model="password2"
                required
                solo
                tabindex="3"
              ></v-text-field>
            </v-flex>
            <p>
              Heedy is ready to create a database with default settings, you just need to give it a starting user.
              For more control on how Heedy is set up, click on the "Server Settings" button.
            </p>
          </v-layout>
        </v-container>
      </v-card-text>

      <v-slide-y-transition>
        <v-card-text v-show="show">
          <div>
            <div class="headline">Server Settings</div>
            <span class="grey--text">Prepare the server's core settings</span>
          </div>

          <v-container>
            <v-layout row wrap>
              <v-flex xs12>
                <h3>Database Location</h3>
                <p>
                  This is the place where heedy will put all its files. It is also the place where settings are saved, and where plugins will be installed.
                  You can choose a different folder by specifying it in the heedy command - this field is readonly.
                </p>
                <v-text-field
                  :placeholder="directoryDefault"
                  v-model.trim="directory"
                  readonly
                  outlined
                ></v-text-field>
              </v-flex>

              <v-flex xs12>
                <h3>Host & Port</h3>
                <p>
                  The main host and port on which to run the server. You should leave the host blank
                  if you want to make Heedy accessible from your phone or other devices on the network.
                  If you want to run Heedy in local mode, so that only things running on the same computer
                  as the server can access it, you can use "localhost".
                </p>
              </v-flex>
              <v-flex sm8 xs12>
                <v-text-field
                  label="Host"
                  :placeholder="hostDefault"
                  v-model.trim="host"
                  required
                  solo
                  tabindex="5"
                ></v-text-field>
              </v-flex>
              <v-flex sm4 xs12>
                <v-text-field
                  label="Port"
                  type="number"
                  :placeholder="portDefault"
                  v-model.number="port"
                  required
                  solo
                  tabindex="6"
                ></v-text-field>
              </v-flex>

              <!--
        <v-flex xs12>
        <h3>HTTPS</h3>
        <p>When accessing heedy over the internet, it is very important to 
          have an encrypted app, so that others can't see your info and passwords.
          If you have a domain name, heedy can automatically set up https for you using Let's Encrypt.</p>
          <v-radio-group v-model="tls" >
      <v-radio
        label="No Encryption"
        value="none"
        tabindex="7"
      ></v-radio>
      <v-radio
        label="Use Let's Encrypt"
        value="letsencrypt"
        tabindex="8"
      ></v-radio>
      <v-radio
        label="Custom"
        value="custom"
        tabindex="9"
      ></v-radio>
    </v-radio-group>
        </v-flex>
              -->
            </v-layout>
          </v-container>
        </v-card-text>
      </v-slide-y-transition>
      <v-card-actions>
        <v-btn text @click="show = !show" tabindex="11">
          <v-icon>{{ show ? 'keyboard_arrow_up' : 'keyboard_arrow_down' }}</v-icon>Server Settings
        </v-btn>
        <v-spacer></v-spacer>
        <v-btn color="info" type="submit" tabindex="10" :loading="loading">Create Database</v-btn>
      </v-card-actions>
    </v-form>
  </v-card>
</template>

<script>
import api from "../../api.mjs";

export default {
  data: () => ({
    show: false,
    directoryDefault: installDirectory,
    directory: installDirectory,
    hostDefault: configuration["host"],
    host: configuration["host"],
    portDefault: configuration["port"].toString(),
    port: configuration["port"].toString(),
    tls: "none",
    username: "",
    password1: "",
    password2: "",
    alert: "",
    success: "",
    loading: false
  }),
  methods: {
    submit: async function(event) {
      event.preventDefault();
      if (this.loading) {
        return;
      }
      this.loading = true;
      this.alert = "";
      window.scrollTo({
        top: 0,
        left: 0,
        behavior: "smooth"
      });
      if (this.username === "") {
        this.alert = "A username is required";
        this.loading = false;
        return;
      }
      if (this.password1 != this.password2) {
        this.alert = "The passwords do not match";
        this.loading = false;
        return;
      }
      if (this.password1 === "") {
        this.alert = "A password is required";
        this.loading = false;
        return;
      }
      let port = parseInt(this.port, 10);
      if (isNaN(port)) {
        this.alert = "The port must be a number";
        this.loading = false;
        return;
      }
      // Generate the query used to create the user.
      let query = {
        user: {
          username: this.username,
          password: this.password1
        },
        config: {
          host: this.host,
          port: port
        }
      };

      // Only add configuration options which have been changed
      if (this.directory !== this.directoryDefault) {
        query.directory = this.directory;
      }

      let result = await fetch("/setup", {
        method: "POST",
        headers: {
          Accept: "application/json",
          "Content-Type": "application/json"
        },
        body: JSON.stringify(query)
      }).catch(error => console.error(error));

      if (result.status != 200) {
        this.alert = (await result.json())["error_description"];
        this.loading = false;
        return;
      }

      let furl = "/auth/token";
      if (this.host != this.hostDefault || this.port != this.portDefault) {
        window.location.href = "http://" + this.host + ":" + this.portDefault;
      }

      // The setup went with defaults, so log in

      this.success = "Database created! Waiting for heedy to restart...";

      function sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
      }

      let i = 0;
      let isok = false;
      do {
        i += 1;
        let res = await api(
          "POST",
          "/auth/token",
          {
            grant_type: "password",
            username: this.username,
            password: this.password1
          },
          null,
          false
        );
        isok = res.response.ok;
        console.log(res);
        if (!isok) {
          await sleep(100);
        }
      } while (i < 5 && !isok);

      // We don't actually care about the result - we just wanted the cookie. Now redirect
      window.location.href = window.location.href.split("setup/")[0];
    }
  }
};
</script>
