#The ConnectorDB nginx configuration

#Redirect all http traffic to https
server {
	listen 80 default_server;
	listen [::]:80 default_server ipv6only=on;

	server_name connectordb.com;
	return 301 https://connectordb.com$request_uri;
}

#The https server serves from a jekyll website, and shows an "app is down" message if
#it can't connect to 
server {
	listen 443;
	server_name connectordb.com;

	root /www/connectordb.com/static/;
	index index.html index.htm;

	error_page 404 /static/404/index.html;
	error_page 502 /static/404/index.html;

	ssl on;
	ssl_certificate /etc/nginx/ssl/connectordb.com.crt;
	ssl_certificate_key /etc/nginx/ssl/connectordb.com.key;

	ssl_session_timeout 5m;

	ssl_protocols TLSv1 TLSv1.1 TLSv1.2; # don't use SSLv3 ref: POODLE
	ssl_ciphers "HIGH:!aNULL:!MD5 or HIGH:!aNULL:!MD5:!3DES";
	ssl_prefer_server_ciphers on;

	location /static/ {
		rewrite ^/static/?(.*)$ /$1 break;
		try_files $uri $uri/ =404;
	}

	location / {
		proxy_pass http://127.0.0.1:8080;
		proxy_set_header Host $host;
		proxy_set_header X-Real-IP $remote_addr;
		proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
	}
}
