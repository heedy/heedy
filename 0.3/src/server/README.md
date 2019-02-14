Server is the full web-facing interface of ConnectorDB.
- webcore: The core underlying code for the entire interface, which includes authentication and logging handlers
- restapi: The REST api for ConnectorDB
- website: The core website and app handling


server.go initializes all of these and sets them up. Look for the individual routers in `router.go` of the rest/webstie subdirectories
