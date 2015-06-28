#The absolute simplest test - each user performs a ping each second
from locust_cdb import *

class PingTask(TaskSet):

	@task
	def ping(self):
		self.client.get("?q=this",auth=("test","test"))



class PingUser(BenchLocust):
	task_set = PingTask
	min_wait = 500
	max_wait = 1500
