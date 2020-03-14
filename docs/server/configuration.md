# Configuring Heedy

If you are an admin of your heedy instance, you can perform all server maintenance
and apply configuration changes directly from the "Settings" UI. Heedy will automatically perform a full database backup when restarting on each configuration change made from the web UI,
so if the server fails to restart for any reason, any changes you made will be automatically reverted.

## Installing Plugins

You can install a plugin using the web UI by uploading a zip file containing the plugin folder.
![Plugin Upload UI](./plugin_upload.png)

You will be prompted to restart heedy to activate the plugin. 
It is highly recommended that you check the box for a database backup whenever installing new plugins which might modify your database. That way, if there are issues, heedy can revert the entire update, including any changes to the database.

### Manually Installing Plugins

You can also manually upload a plugin by extracting the zip file, and placing the plugin folder in the `plugins` subdirectory of your database folder.

That is, suppose your heedy database folder is at `./myheedy`. Then installing the plugin `myplugin` can be achieved as follows using bash:
```
unzip myplugin-1.0.0.zip
mkdir ./myheedy/plugins
mv myplugin ./myheedy/plugins/
```

You will then need to enable the plugin by modifying your heedy.conf `vim ./myheedy/heedy.conf` to add `myplugin` to the active plugins array:
```
active_plugins = ["fitbit","notebook","myplugin"]
```

You will need to restart heedy for the changes to take effect.


## Default Configuration

Your heedy.conf is simply overriding the configuration options defined in the plugins you installed, as well as the core built-in settings.
You can modify most of these options
by setting their values in your heedy.conf.

The built-in configuration, shown here, contains a lot of heedy's internals, including the configurations for the plugins that are built into heedy by default.
For this reason, unless you know what you are doing, it is recommended that you leave most options at their default values.

```eval_rst
.. literalinclude:: ../../assets/heedy.conf
    :language: javascript
```