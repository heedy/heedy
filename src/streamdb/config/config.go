package config

/**

This file provides the main configuration system for ConnectorDB.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
	"flag"
	"streamdb"
)

var (
	Nodetype            = flag.String("node.type", "master", "Whether this node should be considered the 'master' or a 'slave'")
	RedisConnection     = flag.String("redis.connection_uri", "", "The redis connection string")
	MessageConnection   = flag.String("gnatsd.connection_uri", "", "The gnatsd connection string")
	DatabaseConnection  = flag.String("database.cxn_string", "", "The database's connection string")
	WebPort             = flag.Int("web.portnum", 8080, "The port to serve the website on")
	WebInterfaceEnabled = flag.Bool("web.http.enabled", true, "Should the http website be run?")
	WebApiEnabled       = flag.Bool("web.api.enabled", true, "Should the web api be on?")
	WebApiKey           = flag.String("web.http.key", "", "The api key for the web interface")
	ApiApiKey           = flag.String("web.api.key", "", "The api key for the web interface")
	DaisyEnabled        = flag.Bool("web.daisy.enabled", false, "Turn on the daisy components of the web interface")
)
