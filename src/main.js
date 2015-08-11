var app = require('app');  // Module to control application life.
var BrowserWindow = require('browser-window');  // Module to create native browser window.
var Menu = require('menu');
var Tray = require('tray');

// Report crashes to our server.
require('crash-reporter').start();

// Keep a global reference of the window object, if you don't, the window will
// be closed automatically when the javascript object is GCed.
var mainWindow = null;
var appIcon = null;

// Quit when all windows are closed.

app.on('window-all-closed', function() {
  /**if (process.platform != 'darwin') {
    app.quit();
}**/
  return false;
});


// This method will be called when Electron has done everything
// initialization and ready for creating browser windows.
app.on('ready', function() {
  // Create the browser window.
  mainWindow = new BrowserWindow({width: 800, height: 600});

  // and load the index.html of the app.
  mainWindow.loadUrl('file://' + __dirname + '/index.html');

  // Open the devtools.
  mainWindow.openDevTools();

  // Emitted when the window is closed.
  mainWindow.on('closed', function() {
    // Dereference the window object, usually you would store windows
    // in an array if your app supports multi windows, this is the time
    // when you should delete the corresponding element.
    mainWindow = null;
  });

    appIcon = new Tray(__dirname + '/ic_launcher.png');
    var contextMenu = Menu.buildFromTemplate([
      { label: 'Options', type: 'normal' },
      { label: 'Log Data', type: 'checkbox', checked: true },
      { label: 'Quit', type: 'normal' }
    ]);
    appIcon.setToolTip('ConnectorDB Logging.');
    appIcon.setContextMenu(contextMenu);

});
