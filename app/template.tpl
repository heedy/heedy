{{define "app"}}
<!doctype html>
<html class="no-js" lang="">
    <head>
        <meta charset="utf-8">
        <meta http-equiv="x-ua-compatible" content="ie=edge">
        <title>ConnectorDB</title>
        <meta name="description" content="">
        <meta name="viewport" content="width=device-width, initial-scale=1">

        <!--<link rel="apple-touch-icon" href="apple-touch-icon.png">-->
        <!-- Place favicon.ico in the root directory -->

        <link rel="stylesheet" href="/app/css/normalize.css">

        <link rel="stylesheet" href="/app/css/bootstrap.min.css">
        <link rel="stylesheet" href="/app/css/main.css">
        <link rel="stylesheet" href="/app/css/leaflet.css">
        <!--<link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
        Switching to inline styles, so that the webapp works even with no internet.

        The annoying thing is that it needs to load the fonts from local url, and we don't know the exact url used

The above is taken directly from https://google.github.io/material-design-icons/
      -->
      <style>


      /*The roboto font @import url(https://fonts.googleapis.com/css?family=Roboto:300); */
      /* roboto-300 - latin */
@font-face {
  font-family: 'Roboto';
  font-style: normal;
  font-weight: 300;
  src: local('Roboto Light'), local('Roboto-Light'),
       url('/app/fonts/roboto-v15-latin-300.woff2') format('woff2'), /* Chrome 26+, Opera 23+, Firefox 39+ */
       url('/app/fonts/roboto-v15-latin-300.woff') format('woff'); /* Chrome 6+, Firefox 3.6+, IE 9+, Safari 5.1+ */
}


/* Material Icons */
      @font-face {
        font-family: 'Material Icons';
        font-style: normal;
        font-weight: 400;
        src: url(/app/MaterialIcons-Regular.eot); /* For IE6-8 */
        src: local('Material Icons'),
             local('MaterialIcons-Regular'),
             url(/app/fonts/MaterialIcons-Regular.woff2) format('woff2'),
             url(/app/fonts/MaterialIcons-Regular.woff) format('woff'),
             url(/app/fonts/MaterialIcons-Regular.ttf) format('truetype');
      }

      .material-icons {
        font-family: 'Material Icons';
        font-weight: normal;
        font-style: normal;
        font-size: 24px;  /* Preferred icon size */
        display: inline-block;
        line-height: 1;
        text-transform: none;
        letter-spacing: normal;
        word-wrap: normal;
        white-space: nowrap;
        direction: ltr;

        /* Support for all WebKit browsers. */
        -webkit-font-smoothing: antialiased;
        /* Support for Safari and Chrome. */
        text-rendering: optimizeLegibility;

        /* Support for Firefox. */
        -moz-osx-font-smoothing: grayscale;

        /* Support for IE. */
        font-feature-settings: 'liga';
      }
      </style>




        <script src="/app/js/modernizr-2.8.3.min.js"></script>
    </head>
    <body>
        <div id="app"></div>
        <!-- Used for obtaining IP https://github.com/diafygi/webrtc-ips -->
        <iframe id="iframe" sandbox="allow-same-origin" style="display: none"></iframe>
        <script>
        // We define the site URL as a global variable
        var SiteURL = "{{.SiteURL}}";
        var ConnectorDBVersion = "{{ Version }}";
        </script>
        <script src="/app/bundle.js" type="text/javascript"></script>
        <script>
          App.run({{json .}});
        </script>
</body>
</html>
{{end}}
