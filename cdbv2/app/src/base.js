console.log("test");

Vue.component("user", function(resolve, reject) {
  console.log("Running user component");
  dynamicImport("./js/user.jsm").then(resolve);
});
