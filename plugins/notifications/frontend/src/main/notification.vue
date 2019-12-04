<template>
  <v-alert
    :type="n.type.length>0?n.type:'info'"
    :border="small?undefined:'left'"
    :colored-border="!small"
    dismissible
    :dense="small"
    :outlined="small"
    prominent
    :elevation="1"
    @input="del"
    style="background-color: #fdfdfd !important;"
  >
    <v-row>
      <v-col class="grow">
        <h3 :style="{'padding-top': description.length>0? '10px':'0'}">
          <router-link :to="linkpath" v-if="showlink">{{ n.title }}</router-link>
          <span v-else>{{ n.title }}</span>
        </h3>
        <span v-if="description.length>0" v-html="description" style="padding-top: 5px"></span>
      </v-col>
      <v-col class="shrink">
        <v-btn v-for="(v,i) in n.actions" :key="i" outlined @click="linkTo(v)">
          <v-icon v-if="v.icon!=''" left>{{ v.icon }}</v-icon>
          {{ v.title }}
        </v-btn>
      </v-col>
    </v-row>
  </v-alert>
</template>
<script>
import { md } from "../../dist/markdown-it.mjs";
export default {
  props: {
    n: Object,
    link: {
      type: Boolean,
      default: false
    },
    small: {
      type: Boolean,
      default: false
    },
    seen: {
      type: Boolean,
      default: false
    }
  },
  computed: {
    description() {
      if (this.n.description.length == 0) {
        return "";
      }
      let r = md.render(this.n.description);
      // TODO: cache instead of rendering each time
      return r;
    },
    showlink() {
      if (!this.link) return false;
      if (this.n.app !== undefined) return true;
      return false;
    },
    linkpath() {
      console.log(this.n);
      if (this.n.object !== undefined) return `/objects/${this.n.object}`;
      if (this.n.app !== undefined) return `/apps/${this.n.app}`;
      return `/users/${this.n.user}`;
    }
  },
  methods: {
    del(v) {
      let nq = { key: this.n.key };
      if (this.n.object !== undefined) {
        nq.object = this.n.object;
      } else if (this.n.app !== undefined) {
        nq.app = this.n.app;
      } else {
        nq.user = this.n.user;
      }
      this.$store.dispatch("deleteNotification", nq);
    },
    linkTo(v) {
      let url = v.href;
      if (url.startsWith("#")) {
        url = location.href.split("#")[0] + url;
        if (!v.new_window) {
          this.$router.push(v.href.substr(1));
          return;
        }
      } else if (v.href.startsWith("/")) {
        url = location.href.split("#")[0] + v.href.substr(1);
      }
      if (v.new_window) {
        window.open(url, "_blank");
      } else {
        location.href = url;
      }
    }
  },
  watch: {
    n(newN) {
      if (this.seen && !newN.seen) {
        let nq = { key: newN.key };
        if (newN.object !== undefined) {
          nq.object = newN.object;
        } else if (newN.app !== undefined) {
          nq.app = newN.app;
        } else {
          nq.user = newN.user;
        }
        this.$store.dispatch("updateNotification", {
          n: nq,
          u: { seen: true }
        });
      }
    }
  },
  created() {
    if (this.seen && !this.n.seen) {
      let nq = { key: this.n.key };
      if (this.n.object !== undefined) {
        nq.object = this.n.object;
      } else if (this.n.app !== undefined) {
        nq.app = this.n.app;
      } else {
        nq.user = this.n.user;
      }
      this.$store.dispatch("updateNotification", {
        n: nq,
        u: { seen: true }
      });
    }
  }
};
</script>