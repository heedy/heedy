from locust import HttpLocust, TaskSet,task
from requests.auth import HTTPBasicAuth

import connectordb

#Replaces the initializer for ConnectorDB with a custom one which uses the locust client
class ConnectorDBLocust(connectordb.ConnectorDB):
	def __init__(self,client,user,password,ud=None):
		self.url = "/api/v1/"

		auth = HTTPBasicAuth(user,password)
		self.r = client
		self.r.auth = auth
		self.r.headers.update({'content-type': 'application/json'})
		if (ud is None):
			ud = user
		connectordb.Device.__init__(self,self,ud)

class BenchLocust(HttpLocust):
	host = "http://192.168.137.21:8000"
