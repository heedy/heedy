package config

/**

This file provides the main configuration tool for ConnectorDB.

Copyright 2015 - Joseph Lewis <joseph@josephlewis.net>
                 Daniel Kumor <rdkumor@gmail.com>

All Rights Reserved

**/

import (
    "streamdb"
    )

type Environment struct {
	Streamdb *streamdb.Database
}

func InitTool(env *Environment) {

}

func CreateAdminTool(env *Environment) {

}
