#!/bin/bash

chmod +x .tmx

INSTALLDIR=$(pwd)

#This script is run on a just-initialized server. It needs root access.

#First, update the shit out of everything
apt-get -y update
apt-get -y upgrade
apt-get -y autoremove  #Sometimes kernels are not removed automatically

#Next, get all the necessary libraries
apt-get -y install tmux postgresql build-essential redis-server git nginx htop wget
apt-get -y install python-nose python-apsw python-coverage python-pip

#We don't want the servers to start on boot (except for nginx)
systemctl disable postrgresql.service
systemctl stop postgresql.service

#Replace {{CONNECTORDB_DIR}} with the install directory
find . -type f -print0 | xargs -0 sed -i "s@{{CONNECTORDB_DIR}}@${INSTALLDIR}@g"

#Okay, we now put all files where they belong
chmod -R 000 ssl

#Alright, now set up nginx
if [ -f "/etc/nginx/sites-enabled/default" ];
then
    rm /etc/nginx/sites-enabled/default
fi
mv ./nginx_config /etc/nginx/sites-available/connectordb
ln -s /etc/nginx/sites-available/connectordb /etc/nginx/sites-enabled/connectordb

sudo systemctl restart nginx.service

#And now, install a recent version of golang
mkdir tmp
cd tmp
wget https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz
cd ..
rm -rf tmp

#Now clone the database - needs auth
git clone https://github.com/dkumor/connectordb.git

#And finally, set up the python module. This installs the deps for python module
cd connectordb/src/clients/python
python setup.py install
cd ~

#aaaand we're done
echo "Finished"
