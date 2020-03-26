import Header from "./main/header.vue";

function setup(frontend) {
  frontend.objects.addType({
    type: "dashboard",
    title: "Dashboard",
    list_title: "Dashboards",
    icon: "dashboard"
  });
  frontend.objects.addComponent({
    component: Header,
    type: "dashboard",
    key: "header"
  });

  if (frontend.info.user != null) {
    frontend.objects.addCreator({
      key: "dashboard",
      title: "Dashboard",
      description: "Display data from multiple sources",
      icon: "dashboard",
      fn: async () => {
        let res = await frontend.rest("POST", "/api/objects", {
          name: "My Dashboard",
          type: "dashboard"
        });
        if (res.response.ok) {
          frontend.router.push({ path: `/objects/${res.data.id}` });
        } else {
          frontend.store.dispatch("errnotify", res.data);
        }
      }
    });
  }
}

export default setup;
