/*
  This component renders the main app theme (./theme/Theme.js), and shows the page relevant to the routing shown in the URL.

  App represents the expected routing of ConnectorDB, so that we can use the frontend as a single-page application.
  ConnectorDB's web frontend uses a /user/device/stream url handling. All of these are redirected to the same code
  by the frontend. This means that any part of the app could be queried.
  We use the react-router package to manage the urls, and to run the correct javascript for each route.
  All of these represent urls that are directly recognized by the ConnectorDB server.

  The App component is initialized in index.js. It is react's main entry point.
*/

import React, { Component, PropTypes } from "react";
import { connect } from "react-redux";

import { Route } from "react-router";

import Theme from "./theme/Theme";

import MainPage from "./MainPage";
import User from "./User";
import Device from "./Device";
import Stream from "./Stream";

// While the logout url removes all cookies, the frontend uses a special component to do further cleanup,
// since we save a lot of resources offline (so that the app can be used even without internet connection) that
// should be deleted on logout.
import Logout from "./Logout";

const App = () => (
  <Theme>
    <Route exact path="/" component={MainPage} />
    <Route exact path="/logout" component={Logout} />
    <Route exact path="/:user" component={User} />
    <Route exact path="/:user/:device" component={Device} />
    <Route exact path="/:user/:device/:stream" component={Stream} />
  </Theme>
);

export default App;
