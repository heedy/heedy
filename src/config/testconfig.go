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
	c.Sql = &SQLService{
		Type: "postgres",
		Service: Service{
			Hostname: "localhost",
			Port:     52592,
			Username: "postgres",
			Password: "sqlpassword",
			Enabled:  true,
		},
	}

	c.BatchSize = 250
	c.ChunkSize = 1

	// Time out the cache in one second
	c.CacheTimeout = 1000

	// The debug log level shows ALL the messages :)
	c.LogLevel = "debug"

	return *c
}()

// TestUser is the user generated for the testing configuration
var TestUser = &UserMaker{
	Name:     "test",
	Email:    "test@localhost",
	Password: "test",
	Role:     "admin",
}

//TestOptions is the options of tests
var TestOptions = TestConfiguration.Options()
