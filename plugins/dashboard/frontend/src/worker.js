import DashboardWorkerInjector from "./worker/injector.js";

function setup(wkr) {
  wkr.inject("dashboard", new DashboardWorkerInjector(wkr));
}

export default setup;
