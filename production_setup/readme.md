Setting Up Production Server
==============================

The instructions were created during initial setup of connectordb.com server using a DigitalOcean droplet
and Namecheap DNS.

I am assuming this is a clean droplet. [Good instructions are here](https://www.digitalocean.com/community/tutorials/initial-server-setup-with-ubuntu-14-04)

Starting off:
```
apt-get update
apt-get dist-upgrade
apt-get install nginx base-devel git golang golang-go.tools
```

Then create the user:
```
adduser support
gpasswd -a sudo support
```

log in as the user, and copy the dotfiles from this repo to the home directory.
Then, disable root login (`PermitRootLogin no`):
```
sudo vim /etc/ssh/sshd_config
sudo service ssh restart
```

Now make sure you can log in to the server as `support` user

# Static Website

Static pages, such as the 404 page and the greeting screen are served at `connectordb.com/static/*`.
To set this up:

The static website uses jekyll. We need to install it:
```
sudo apt-get install rubygems ruby-dev
sudo gem install jekyll
sudo gem install typogruby
sudo gem install python-pygments
```

Now that jekyll is installed, we set up the static pages location

```
sudo mkdir /www
cd /www
sudo mkdir connectordb.com
cd connectordb.com
sudo mkdir static
sudo chmod 777 static
```

The chmod is needed so nginx can read the files

Next, set up the git repo for the site:

```
cd ~
git init --bare connectordb_static.git

```

Copy the files from static_site_hooks in repo to `./connectordb_static/hooks`. The hooks auto-deploy the website on push.

Lastly, push a jekyll site to `support@connectordb.com:connectordb_static.git` with at least a 404 page.

# NGINX

Nginx is used as an SSL proxy for the web app at port 8080, and also serves the jekyll website set up above at `/static/`.

To start off, copy `nginx/connectordb.com` from this folder into /etc/nginx/sites-available

Then link it to sites-enabled:
```
cd /etc/nginx/sites-enabled
sudo ln -s /etc/nginx/sites-available/connectordb.com ./connectordb.com
```


### SSL Certificate

Namecheap PositiveSSL is the one we used before, so instructions are for that.

[These instructions were followed](https://www.digitalocean.com/community/tutorials/how-to-install-an-ssl-certificate-from-a-commercial-certificate-authority)

```
cd ~
openssl req -newkey rsa:2048 -nodes -keyout connectordb.com.key -out connectordb.com.csr
```

The csr was sent to Comodo for issuing, and a large chain of files was returned:

```
AddTrustExternalCARoot.crt
COMODORSAAddTrustCA.crt
COMODORSADomainValidationSecureServerCA.crt
connectordb_com.crt
```

This command created the pem key:

```
cat connectordb_com.crt COMODORSADomainValidationSecureServerCA.crt COMODORSAAddTrustCA.crt AddTrustExternalCARoot.crt > connectordb.com.crt
```

This pem file needs to be put in the correct folder:

```
cd ~
sudo mkdir /etc/nginx/ssl
sudo mv ./connectordb.com.crt /etc/nginx/ssl/
sudo mv ./connectordb.com.key /etc/nginx/ssl/
cd /etc/nginx/ssl
sudo chown root:root ./*
chmod 000 ./*
cd ..
chmod 000 ./ssl
```




### Finishing up...
```
sudo service nginx restart
```

There is about a 99% chance that it won't work right away. The nginx error logs are in `/var/log/nginx`


# ConnectorDB

These instructions show how to set up ConnectorDB to auto-deploy on push

```
cd ~
git init --bare connectordb.git
```

*TODO: This, and the hooks to deploy*
