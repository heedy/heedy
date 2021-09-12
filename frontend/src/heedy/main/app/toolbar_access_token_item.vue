<template>
  <v-dialog v-model="showkey" width="500">
    <template #activator="{ on: onDialog }">
      <v-tooltip bottom v-if="!isList">
        <template #activator="{ on }">
          <v-btn
            icon
            v-on="{ ...on }"
            @click="
              (e) => {
                onDialog.click(e);
                getKey();
              }
            "
          >
            <v-icon>vpn_key</v-icon>
          </v-btn>
        </template>
        <span>View Access Token</span>
      </v-tooltip>
      <v-list-item
        v-else
        @click="
          (e) => {
            onDialog.click(e);
            getKey();
          }
        "
      >
        <v-list-item-icon>
          <v-icon>vpn_key</v-icon>
        </v-list-item-icon>
        <v-list-item-content>
          <v-list-item-title>View Access Token</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </template>
    <v-card>
      <v-card-title text-center>
        <v-layout row justify-center>
          <v-flex text-center style="padding-top: 15px; padding-bottom: 15px">
            <h3 style="color: #1976d2; padding-bottom: 7px">Access Token</h3>
            <h4>{{ token }}</h4>
            <h6 style="color: gray">Last used: {{ accessed }}</h6>
            <h6 style="color: gray">
              You can reset the token in
              <v-icon>edit</v-icon>
            </h6>
            <v-btn
              rounded
              outlined
              style="margin-top: 15px"
              color="grey"
              @click="showkey = false"
              >Done</v-btn
            >
          </v-flex>
        </v-layout>
      </v-card-title>
    </v-card>
  </v-dialog>
</template>
<script>
import api from "../../../util.mjs";
import Moment from "../../../dist/moment.mjs";
export default {
  props: {
    appid: String,
    isList: Boolean,
  },
  data: () => ({
    showkey: false,
    token: "...",
    accessed: "...",
  }),
  watch: {
    showkey(newv) {
      this.token = "...";
    },
  },
  methods: {
    getKey: async function () {
      console.vlog("Reading access token for", this.appid);
      let result = await api(
        "GET",
        `api/apps/${encodeURIComponent(this.appid)}`,
        {
          token: true,
        }
      );
      if (!result.response.ok) {
        this.token = result.data.error_description;
        return;
      }
      this.token = result.data.access_token;

      if (result.data.last_access_date == null) {
        this.accessed = "never";
      } else {
        this.accessed = Moment(result.data.last_access_date).calendar(null, {
          sameDay: "[Today]",
          nextDay: "[Tomorrow]",
          nextWeek: "dddd",
          lastDay: "[Yesterday]",
          lastWeek: "[Last] dddd",
          sameElse: "DD/MM/YYYY",
        });
      }
      return;
    },
  },
};
</script>
