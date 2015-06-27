#The absolute simplest test - each user performs a ping each second
from locust_cdb import *

class PingTask(TaskSet):
	def on_start(self):
		self.cdb = ConnectorDBLocust(self.client,"test","test","test/user")
	@task
	def ping(self):
		self.cdb.ping()



class PingUser(BenchLocust):
	task_set = PingTask
	min_wait = 500
	max_wait = 1500
