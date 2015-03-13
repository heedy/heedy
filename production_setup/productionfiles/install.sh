#!/bin/bash

INSTALLDIR=/home/connectordb

#This script is run on a just-initialized server. It needs root access.

#First, update the shit out of everything
apt-get update
apt-get dist-upgrade
apt-get autoremove  #Sometimes kernels are not removed automatically

#Next, get all the necessary libraries
apt-get install base-devel postgres redis-server git golang golang-go.tools rubygems nodejs python-pygments ruby-dev

#We don't want the servers to start on boot (except for nginx)
update-rc.d postgres disable
update-rc.d redis-server disable
service postgres stop
service redis-server stop

#Now we install jekyll's dependencies
gem install jekyll
gem install typogruby

#Create the connectordb user (which uses git shell for its login)
useradd -r -d $INSTALLDIR -s /usr/bin/git-shell connectordb


#Replace {{CONNECTORDB_DIR}} with the install directory
find . -type f -print0 | xargs -0 sed -i "s/{{CONNECTORDB_DIR}}/${INSTALLDIR}/g"

#Okay, we now put all files where they belong

#First, the connectordb folder is put in the right place
mv connectordb $INSTALLDIR

#Set permissions so that nginx can read the www directory
chmod -R 700 $INSTALLDIR
chown -R connectordb:connectordb $INSTALLDIR
chmod $INSTALLDIR 755
chmod -R $INSTALLDIR/www 755

#Next put the ssl keys in their spot - they are assumed to be in current directory
mkdir $INSTALLDIR/ssl
mv connectordb.key $INSTALLDIR/ssl/
mv connectordb.crt $INSTALLDIR/ssl/
chmod -R 000 $INSTALLDIR/ssl

#Alright, now set up nginx
if [ -f "/etc/nginx/sites-enabled/default" ];
then
    rm /etc/nginx/sites-enabled/default
fi
mv nginx_config /etc/nginx/sites-available/connectordb
ls -s /etc/nginx/sites-available/connectordb /etc/nginx/sites-enabled/connectordb

service nginx restart

#At this point basics are all done. Now set up git for the public facing site
#and for connectrodb itself
su connectordb <<'EOF'
mkdir -m 700 git
cd git
mkdir tmp
git init --bare public.git
git init --bare connectordb.git
EOF

#Now that the repositories are created, put the build hooks in them
chown connectordb:connectordb post-receive_connectordb
chown connectordb:connectordb post-receive_public
chmod +x post-receive_connectordb
chmod +x post-receive_public
mv post-receive_connectordb $INSTALLDIR/git/connectordb.git/hooks/post-receive
mv post-receive_public $INSTALLDIR/git/public.git/hooks/post-receive

#Lastly, create the encrypted container which will hold the databases


#aaaand we're done
echo "Finished"


