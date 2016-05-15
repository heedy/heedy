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

        <link rel="stylesheet" href="{{.SiteURL}}/app/css/normalize.css">
        <link rel="stylesheet" href="{{.SiteURL}}/app/css/main.css">
        <link href="https://fonts.googleapis.com/icon?family=Material+Icons" rel="stylesheet">
        <script src="{{.SiteURL}}/app/js/modernizr-2.8.3.min.js"></script>
    </head>
    <body>
        <!--[if lt IE 8]>
            <p class="browserupgrade">You are using an <strong>outdated</strong> browser. Please <a href="http://browsehappy.com/">upgrade your browser</a> to improve your experience.</p>
        <![endif]-->
        <div id="app"></div>
        <script>
        // We define the site URL as a global variable
        var SiteURL = "{{.SiteURL}}";
        </script>
        <script src="/app/bundle.js" type="text/javascript"></script>
        <script>
          App.run({{json .}});
        </script>
</body>
</html>
{{end}}
