Setting Up Production Server
==============================

The instructions were created during initial setup of connectordb.com server using a DigitalOcean droplet
and Namecheap DNS.

I am assuming this is a clean droplet. [Good instructions are here](https://www.digitalocean.com/community/tutorials/initial-server-setup-with-ubuntu-14-04)

Starting off:
```
apt-get update
apt-get dist-upgrade
```

Then create the support user:
```
adduser support
gpasswd -a sudo support
```

log in as the support user, and copy the dotfiles from this repo to the home directory.
Then, disable root login (`PermitRootLogin no`):
```
sudo vim /etc/ssh/sshd_config
sudo service ssh restart
```

Now make sure you can log in to the server as `support` user. From now on, instructions are run as support user.

### SSL Certificate

Namecheap PositiveSSL is the one we used before, so instructions are for that.

[These instructions were followed](https://www.digitalocean.com/community/tutorials/how-to-install-an-ssl-certificate-from-a-commercial-certificate-authority)

```
cd ~
openssl req -newkey rsa:2048 -nodes -keyout connectordb.key -out connectordb.com.csr
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
cat connectordb_com.crt COMODORSADomainValidationSecureServerCA.crt COMODORSAAddTrustCA.crt AddTrustExternalCARoot.crt > connectordb.crt
```

This pem file needs to be put in the correct folder:




# Install Script

After having the `connectordb.crt` and `connectordb.key` files, you need to copy the productionfiles directory to the server, and put the two ssl files in it.

Then, you can run the install script (NOTE: Install script was not yet tested!!!):
```
mv connectordb.key productionfiles/
mv connectordb.crt productionfiles/
cd productionfiles
chmod +x install.sh
sudo ./install.sh
```

After the install script finishes, you will have the prerequisites for running connectordb installed.
Check to make sure that everything is working by navigating to http://{{nameHere}}.com.

Two things should happen: the http should be redirected to https, and the https should be green (valid).
Furthermore, the page should be a 404 explicitly saying something about connectordb.

At that point, create a password for the connectordb user
```
passwd connectordb
```
and that will allow you to push-to-deploy:
```
connectordb@connectordb.com/git/public.git #Jekyll website at /public . Must have a 404 page.
connectordb@connectordb.com/git/connectordb.git #The connectordb repo - auto-deploys a production repository on push.
```

There is about a chance that something won't work right away. The nginx error logs are in `/var/log/nginx`
