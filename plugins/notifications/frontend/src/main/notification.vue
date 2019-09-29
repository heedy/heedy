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
    <h3 :style="{'padding-top': description.length>0? '10px':'0'}">
      <router-link :to="linkpath" v-if="showlink">{{ n.title }}</router-link>
      <span v-else>{{ n.title }}</span>
    </h3>
    <span v-if="description.length>0" v-html="description" style="padding-top: 5px"></span>
  </v-alert>
</template>
<script>
import { md } from "../../dist.mjs";
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
      if (this.n.connection !== undefined) return true;
      return false;
    },
    linkpath() {
      console.log(this.n);
      if (this.n.source !== undefined) return `/sources/${this.n.source}`;
      if (this.n.connection !== undefined)
        return `/connections/${this.n.connection}`;
      return `/users/${this.n.user}`;
    }
  },
  methods: {
    del(v) {
      let nq = { key: this.n.key };
      if (this.n.source !== undefined) {
        nq.source = this.n.source;
      } else if (this.n.connection !== undefined) {
        nq.connection = this.n.connection;
      } else {
        nq.user = this.n.user;
      }
      this.$store.dispatch("deleteNotification", nq);
    }
  },
  watch: {
    n(newN) {
      if (this.seen && !newN.seen) {
        let nq = { key: newN.key };
        if (newN.source !== undefined) {
          nq.source = newN.source;
        } else if (newN.connection !== undefined) {
          nq.connection = newN.connection;
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
      if (this.n.source !== undefined) {
        nq.source = this.n.source;
      } else if (this.n.connection !== undefined) {
        nq.connection = this.n.connection;
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