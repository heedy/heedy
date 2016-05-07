{{define "header"}}
<!doctype html>
<html class="no-js" lang="">
    <head>
        <meta charset="utf-8">
        <meta http-equiv="x-ua-compatible" content="ie=edge">
        <title></title>
        <meta name="description" content="">
        <meta name="viewport" content="width=device-width, initial-scale=1">

        <!--<link rel="apple-touch-icon" href="apple-touch-icon.png">-->
        <!-- Place favicon.ico in the root directory -->

        <link rel="stylesheet" href="app/css/normalize.css">
        <link rel="stylesheet" href="app/css/main.css">
        <script src="app/js/modernizr-2.8.3.min.js"></script>
    </head>
    <body>
        <!--[if lt IE 8]>
            <p class="browserupgrade">You are using an <strong>outdated</strong> browser. Please <a href="http://browsehappy.com/">upgrade your browser</a> to improve your experience.</p>
        <![endif]-->
{{end}}
{{define "footer"}}
<div id="footer">
    ConnectorDB {{ .Version }}<br/>
    &copy; 2016 The <a href="https://connectordb.github.io" >ConnectorDB</a> contributors.
</div>
</body>
</html>
{{end}}
