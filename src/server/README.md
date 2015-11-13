Server is the full web-facing interface of ConnectorDB.
- webcore: The core underlying code for the entire interface, which includes authentication and logging handlers
- restapi: The REST api for ConnectorDB

This code all assumes that in the same directory as the binary connectordb there are two folders:
- www: The website to show when not authenticated/logged in
	- Assumed to have the following pages:
		- index.html: The main webpage to show when not logged in
		- login.html: A login page to show when attempting to access resources
		- join.html: A form which is used to create new users
		- 404.html: Page to show upon a 404 error
- app: The app which is shown to logged in users
