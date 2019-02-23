<template>
    <v-card color="rgb(245,245,250)">

        <v-card-title primary-title>
          <div>
            <div class="headline">Create Your User</div>
            <span class="grey--text">The first user is created as a database administrator.</span>
          </div>
        </v-card-title>
      <v-form v-model="valid">
        <v-card-text>
          <v-container>
          <v-layout row wrap>
            <v-flex xs12>
              <h3>Username</h3>
              <v-text-field
            label="Username"
            placeholder="admin"
            value=""
            required
            solo
          ></v-text-field>
              </v-flex>
              <v-flex xs12>
              <h3>Password</h3>
              <v-layout row wrap>
          <v-flex md6 xs12 >
          <v-text-field
            label="Password"
            placeholder="Password"
            type="password"
            required
            solo
          ></v-text-field>
          </v-flex>
          <v-flex md6 xs12 >
          <v-text-field
            label="Repeat Password"
            placeholder="Repeat Password"
            type="password"
            required
            solo
          ></v-text-field>
          </v-flex>
        </v-layout>
        
              </v-flex>
              <p>ConnectorDB is ready to create a database with default settings, you just need to give it a starting user.
                For more control on how ConnectorDB is set up, click on the "Server Settings" button.
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
          <v-text-field
            label="Database Location"
            :placeholder="directoryDefault"
            :value="directory"
            required
            solo
          ></v-text-field>
          <p>This is the place to put all files needed to run ConnectorDB. It is also the place where settings are saved, and where plugins will be installed.</p>
        </v-flex>

        <v-flex xs12 >
        <h3>Host & Port</h3>
        
        <v-layout row wrap>
          <v-flex sm6 xs12 >
          <v-text-field
            label="Host"
            :placeholder="hostDefault"
            :value="host"
            required
            solo
          ></v-text-field>
          </v-flex>
          <v-flex sm6 xs12 >
          <v-text-field
            label="Port"
            :placeholder="portDefault"
            :value="port"
            required
            solo
          ></v-text-field>
          </v-flex>
        </v-layout>
        <p>The main host and port on which to run the server. You should leave the host blank
            if you want to make ConnectorDB accessible from your phone or other devices on the network.
            If you want to run ConnectorDB in local mode, so that only things running on the same computer
            as the server can access it, you can use "localhost".
            </p>
          
            
        </v-flex>

        <v-flex xs12>
        <h3>HTTPS</h3>
          <v-radio-group v-model="tls">
      <v-radio
        label="Self-Signed Certificate"
        value="selfsigned"
      ></v-radio>
      <v-radio
        label="Use Let's Encrypt Certificate"
        value="letsencrypt"
      ></v-radio>
      <v-radio
        label="Custom"
        value="custom"
      ></v-radio>
    </v-radio-group>
          <p>
            The main port is always https,
            so you must also choose how to encrypt the port. If you are running it at home, the default
            self-signed certificate is good enough. If you are running it on the web, then you should choose
            let's encrypt to generate valid certificates. Finally, you are also free to provide your own 
            encryption keys.
          </p>
        </v-flex>


        <v-flex xs12>
        <h3>HTTP</h3>
          <v-layout row wrap>
      <v-flex xs12 sm4>
        <v-checkbox input-value="true" value label="Enable HTTP port"></v-checkbox>
      </v-flex>
      <v-flex xs12 sm8>
        <v-text-field
            label="Port"
            :placeholder="portDefault"
            :value="port"
            required
            solo
          ></v-text-field>
      </v-flex>
    </v-layout>
          <p>
            If running at home, your browser will give error messages when trying to connect to the https port,
            because it does not recognize self-signed certificates. ConnectorDB therefore also allows you to expose
            an unencrypted port. Beware, though - anyone on your network can read your passwords and data when using this port!
          </p>
        </v-flex>
  <!--
        <v-flex
          xs12
          md4
        >
          <v-text-field
            v-model="lastname"
            :rules="nameRules"
            :counter="10"
            label="Last name"
            required
          ></v-text-field>
        </v-flex>

        <v-flex
          xs12
          md4
        >
          <v-text-field
            v-model="email"
            :rules="emailRules"
            label="E-mail"
            required
          ></v-text-field>
        </v-flex>-->
      </v-layout>
    </v-container>
  
          </v-card-text>
        </v-slide-y-transition>
        </v-form>

        <v-card-actions>
            <v-btn flat @click="show = !show">
            <v-icon>{{ show ? 'keyboard_arrow_up' : 'keyboard_arrow_down' }}</v-icon> Server Settings
          </v-btn>
          <v-spacer></v-spacer>
          <v-btn color="info">Create Database</v-btn>
          
          
        </v-card-actions>
      </v-card>
</template>

<script>
export default {
  data: () => ({
      show: false,
      directoryDefault: installDirectory,
      directory: installDirectory,
      hostDefault: configuration["host"],
      host: configuration["host"],
      portDefault: configuration["port"],
      port: configuration["port"],
      tls: "selfsigned"
    }),
  props: {
  }
};
</script>
