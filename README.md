# Heedy
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/heedy/heedy/Test?label=tests&style=flat-square)

**Note:** *Heedy replaces ConnectorDB, and is currently a pre-alpha. For ConnectorDB 0.3, please go to the releases page.* 

A repository for your quantified-self data, and an extensible analysis engine.

There already exist many apps and fitness trackers that gather and attempt to make sense of your data. Most of these services are isolated - your phone's fitness tracking software knows nothing about your browser's time-tracking extension. Furthermore, each app and service has its own method for downloading data (if they offer raw data at all!), which makes an all-encompassing analysis of life extremely tedious. Heedy offers a self-hosted open-source way to put all of this data together into a single system.

Several existing aggregators already perform many of heedy's functions ([see the list here](https://github.com/woop/awesome-quantified-self#aggregators--dashboards)). However, they are all missing one of two critical components:

1) **Open-source and self-hosted**: Most existing tools are cloud-based, which means that all of your data is on "someone else's computer". While these companies may claim that they will not [sell your data](https://arstechnica.com/information-technology/2017/03/how-isps-can-sell-your-web-history-and-how-to-stop-them/), or won't [turn it over to governments](https://en.wikipedia.org/wiki/Lavabit), they can change their minds (and terms of service) at any time. The only way to guarantee that your data will never be used against you is for it to be on your computer, operated by software you can audit yourself.
2) **Extensible**: Even a system with fantastic visualizations and powerful analysis, like heedy's predecessor (ConnectorDB), has limited utility. This is because it can only perform the analysis the original authors assumed would be useful. While ConnectorDB included a REST API, it was tedious and required lots of computer knowledge to run specialized analysis scripts. Furthermore, any application built upon CDB needed to create its own separate UI. Heedy offers a powerful plugin system - plugin writers can add new integrations, plots, or even modify core functionality with a few lines of python or javascript. A registry is planned, so that users can install plugins with the click of a button.

# Installing

1) Download the executable
2) Run the executable
3) Open your browser to http://localhost:1324


# Plugins

Heedy itself is very limited in scope. Most of its power comes from the plugins that allow you to integrate it with other services. Some plugins worth checking out:

- [fitbit](https://github.com/heedy/heedy-fitbit-plugin) - sync heedy with Fitbit, allowing you to access and analyze your wearables' data.
- [notebook](https://github.com/heedy/heedy-notebook-plugin) - analyze data directly within Heedy with Jupyter notebooks.

# Building

Building heedy requires at least go 1.13 and a recent version of node and npm.

## Release

```
git clone https://github.com/heedy/heedy
cd heedy
make
```

## Debug

```
git clone https://github.com/heedy/heedy
cd heedy
make debug
```

The debug version uses the assets from the `./assets` folder instead of embedding in the executable.

### Watch frontend

To edit the frontend, you will want a debug build and run the following in the frontend folder:
```
npm run debug
```

This will watch all the files, allowing you to see changes by refreshing your browser.

### Verbose Mode

You can see everything heedy does, including all SQL statements and raw http requests by running it in verbose mode:
```
./heedy --verbose
```