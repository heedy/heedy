# Setup & Install

The heedy server comes as a single-file executable, which you can download from the [github releases page](https://github.com/heedy/heedy/releases). All you need to do is run `heedy` without any arguments, and it will guide you through setting up and running a database.
For casual use, *nothing else is needed* - that is, everything should *just work*. The remainder of this document is therefore focused on advanced users, who want more control over their heedy install.

## Python Environment

Since most heedy plugins are written in Python, Heedy attempts to find a valid Python `>=3.7` install automatically while creating a database. It sets this Python up in `heedy.conf`, 
or leaves the configuration blank if no supported interpreter is found. You can see and edit the chosen interpreter in heedy.conf (accessible from the database folder, or from the heedy server settings UI)
by modifying the `python` plugin `path` setting:
```javascript
plugin "python" {

  // Path to the python >=3.7 interpreter to use for python-based plugins.
  path = "/usr/bin/python3"

  // Command-line arguments to pass to pip (pip install mypackage {args}
  // or pip install -r requirements.txt {args}). If using the system python
  // on linux/mac, you will need to add "--user" to avoid permissions failure.
  pip_args = ["--user","--quiet"]
}
```

### Using a venv

All plugin requirements are installed using `pip`. Since heedy uses your system Python by default, it means that it will install all relevant modules *in your system python*.
You can avoid this by creating a venv for heedy:
```bash
python3 -m venv /home/myuser/.heedy_env
```
And then you can set the venv's python interpreter in the heedy configuration, which will install all requirements in the venv:
```javascript
plugin "python" {
  path = "/home/myuser/.heedy_env/bin/python"

  // Remove the --user arg when using venv
  pip_args = ["--quiet"]
}
```

## Custom Database Location
For advanced users, if you want control over your database location, you can tell heedy to create a new database in the `mydb` folder with:
```
heedy create ./mydb
```
You can run this database by calling:
```
heedy run ./mydb
```
or you can start it in the background by running:
```
heedy start ./mydb
```
The database can then be stopped with:
```
heedy stop ./mydb
```

## Putting Heedy Online

While heedy will run without issues on your local network, some integrations and plugins require that heedy is accessible from the internet, and has its own domain name.
Therefore, to take advantage of heedy's full power, you will want to set it up on a webserver. You will need to purchase a domain name from a provider such as [Namecheap](https://namecheap.com),
and link it to a server you control. As an example, the [heedy.org](https://heedy.org) website is hosted using [DigitalOcean](https://digitalocean.com). A $5/month droplet should be sufficient
to run a basic heedy install.

While setting up a domain and server is rather technical, and outside the scope of this documentation, [tutorials are readily available online](https://www.digitalocean.com/community/tutorials/initial-server-setup-with-ubuntu-18-04).

If you already have a domain `mydomain.com`, it is recommended that you set up heedy on the subdomain `heedy.mydomain.com`. 
You should put heedy behind a [caddy](https://caddyserver.com/) server, which will automatically set up https, but nginx and others should also work fine, if you're willing to deal with https details.

Adding the following to your `Caddyfile` is sufficient to configure Caddy for heedy:

```
heedy.mydomain.com {
  reverse_proxy localhost:1324
}
```

You can then run heedy to create your database, and access the setup at `https://heedy.mydomain.com`. While setting up, click the "Server Settings" button to make sure that your host is set to localhost, and the site url is set correctly:

![Heedy Create Settings](./create_settings.png)


If your database was already created, you can achieve the same effect by modifying your heedy configuration, and restarting:
```javascript
// This is the host on which to listen to connections. 
// An empty string means to listen on all interfaces
host = "localhost"

// The port on which to run heedy
port = 1324

// URL at which the server can be accessed. Used for callbacks. If empty,
// this value is populated automatically with the server port and LAN IP.
url = "https://heedy.mydomain.com"
```




## Encrypting your Database

Heedy will be holding very personal data, so you might want to encrypt your database, especially if running on a VPS, where you don't control the server's hard drives.

While heedy does not come with built-in encryption support, if on linux (such as a DigitalOcean droplet), you can use a basic python2 script called [cryptify](https://github.com/dkumor/cryptify), with which you can set up an encrypted password-protected container for you heedy database:
```
sudo apt-get install python-subprocess32 cryptsetup
wget -O cryptify https://raw.githubusercontent.com/dkumor/cryptify/master/cryptify
chmod +x cryptify
```

With this script, you can create a 20GB container for your database, saved as `mydatabase.crypt`, and mounted in the folder `mydatabase` using the following command:
```
./cryptify -i mydatabase.crypt -o mydatabase -s 20000 create
heedy create ./mydatabase/heedydb
```

On future reboots you can decrypt your folder by running the following:
```
./cryptify -i mydatabase.crypt -o mydatabase open
heedy run ./mydatabase/heedydb
```

You can then unmount the folder with
```
./cryptify -i mydatabase.crypt -o mydatabase close
```