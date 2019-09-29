<template>
  <h-page-container noflex>
    <component v-for="c in components" :key="c.key" :is="c.component" :user="user" />
  </h-page-container>
</template>
<script>
import { filterComponents } from "../util.js";
export default {
  props: {
    user: Object
  },
  computed: {
    components() {
      return filterComponents(
        this.$store.state.heedy.user_components,
        { curuser: 1 },
        c => {
          if (c.curuser !== undefined && c.curuser) {
            if (this.$store.state.app.info.user == null) {
              return !c.curuser;
            }
            if (
              this.user.username == this.$store.state.app.info.user.username
            ) {
              return c.curuser;
            }
            return !c.curuser;
          }
          return true;
        }
      );
    }
  }
};
</script>