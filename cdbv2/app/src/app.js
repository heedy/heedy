//import Router from 'vue-router';

// Add the vue router
//Vue.use(Router);

Vue.component("user", function(resolve, reject) {
  console.log("Running user component");
  dynamicImport("./js/user.jsm").then(resolve);
});
