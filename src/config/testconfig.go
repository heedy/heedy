/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package config

//TestConfiguration is the configuratino used when testing the database
var TestConfiguration = func() Configuration {
	c := NewConfiguration()
	c.Redis = Service{
		Hostname: "localhost",
		Port:     6379,
		Password: "redis",
		Enabled:  true,
	}
	c.Nats = Service{
		Hostname: "localhost",
		Port:     4222,
		Username: "connectordb",
		Password: "nats",
		Enabled:  true,
	}
	c.Sql = Service{
		Hostname: "localhost",
		Port:     52592,
		//Username: "connectordb",
		//Password: sqlpassword,
		Enabled: true,
	}

	c.InitialUser = &UserMaker{
		Name:     "test",
		Email:    "test@localhost",
		Password: "test",
		Role:     "admin",
	}

	c.BatchSize = 250
	c.ChunkSize = 1

	// The debug log level shows ALL the messages :)
	c.LogLevel = "debug"

	return *c
}()

//TestOptions is the options of tests
var TestOptions = TestConfiguration.Options()
