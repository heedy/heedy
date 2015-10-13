- connectordb.go is the main file which compiles into the connectordb executable
- util is miscellaneous utilities which are used all over the database
- shell is the implementation of the command line shell
- dbsetup is code that can create start and stop the underlying postgres/redis/nats servers
- config contains the definitions and code associated with the core ConnectorDB configuration including methods to read and write the config files.
- server is the code for the frontend server, it includes the rest api and web app and all their associated resources

- connectordb is the core database code
