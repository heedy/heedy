/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package config

import "fmt"

//TestConfiguration is the configuratino used when testing the database
var TestConfiguration = Configuration{
	Version: 1,
	Redis: Service{
		Hostname: "localhost",
		Port:     6379,
		Password: "redis",
		Enabled:  true,
	},
	Nats: Service{
		Hostname: "localhost",
		Port:     4222,
		Username: "connectordb",
		Password: "nats",
		Enabled:  true,
	},
	Sql: Service{
		Hostname: "localhost",
		Port:     52592,
		//Username: "connectordb",
		//Password: sqlpassword,
		Enabled: true,
	},
	//DBWriter: true,

	//The ConnectorDB frontend server
	Service: Service{
		Hostname: "0.0.0.0",
		Port:     8000,
		Enabled:  false,
	},

	DisallowedNames: []string{"support", "www", "api"},

	//The defaults to use for the batch and chunks
	BatchSize: 250,
	ChunkSize: 1,

	//The initial user created of name test and password test
	InitialUsername:     "test",
	InitialUserEmail:    "test@localhost",
	InitialUserPassword: "test",
}

//TestOptions is the options of tests
var TestOptions = TestConfiguration.Options()

func ExampleConfiguration_IsAllowedEmail() {
	cfg := Configuration{}
	fmt.Printf("%v\n", cfg.IsAllowedEmail("foo@bar.com"))

	cfg.AllowedEmailSuffixes = []string{"bar.com", "baz.com"}

	fmt.Printf("%v\n", cfg.IsAllowedEmail("foo@foo.com"))
	fmt.Printf("%v\n", cfg.IsAllowedEmail("foo@bar.com"))
	fmt.Printf("%v\n", cfg.IsAllowedEmail("foo@baz.com"))
	fmt.Printf("%v\n", cfg.IsAllowedEmail("foo@subdomain.baz.com"))

	// Output:
	// true
	// false
	// true
	// true
	// true
}
