# connectordb
A database that connects stuff


#Golang

Sarama: golang kafka client

go-zookeeper: golang zookeeper client

```bash
go get github.com/Shopify/sarama
get get github.com/samuel/go-zookeeper
```

#Python

*Note: This is code ripped out of python setup, it is temporary to get things running fast.
Will be replaced when we have the time.*

This is for running the python which runs Kafka and Zookeeper. For now.

```bash
pip install kazoo
pip install subprocess32
pip install kafka-python
```

To run kafka:

First download kafka:
```bash
cd config
mkdir /tmp/connectorsetup
wget -P /tmp/connectorsetup "http://apache.lauf-forum.at/kafka/0.8.1.1/kafka_2.10-0.8.1.1.tgz"
tar -zxvf "/tmp/connectorsetup/kafka_2.10-0.8.1.1.tgz" -C /tmp/connectorsetup
mkdir ./bin
find /tmp/connectorsetup/kafka_2.10-0.8.1.1/libs/ -iname "*.jar" -exec mv {} ./jar\;
rm -rf /tmp/connectorsetup
```

After the jar files of kafka are in the `config/bin` directory, you can run zookeeper and kafka:

```bash
cd config
python connect.py -c setup.cfg
```
