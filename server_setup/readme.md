Setting Up a Server
==============================

The instructions were created during initial setup of connectordb.com server using a DigitalOcean droplet
and Namecheap DNS.

I am assuming this is a clean droplet. [Good instructions are here](https://www.digitalocean.com/community/tutorials/initial-server-setup-with-ubuntu-14-04)

Starting off:
```
apt-get update
apt-get upgrade
```

Then create the support user:
```
adduser support
gpasswd -a support sudo
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

After having the `connectordb.crt` and `connectordb.key` files, you need to copy the server_setup directory to the server, and put the two ssl files in the ssl subdirectory

make sure you have tmux installed (or make sure that you stay logged in after copying files)

Then, you can run the install script
```
chmod +x install.sh
sudo ./install.sh
```
The install script will probably need your git auth (since private repo).

Once it finishes,

```bash
cd connectordb
make
cd ..
cp -R connectordb/bin ./bin

./bin/connectordb create database
./bin/connectordb start database
./bin/connectordb run database
```

There is about a 99% chance that something won't work right away. The nginx error logs are in `/var/log/nginx`
