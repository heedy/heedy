#The absolute simplest test - each user performs a ping each second
from locust_cdb import *
import uuid
class StreamWriteTask(TaskSet):
	def on_start(self):
		self.cdb = ConnectorDBLocust(self.client,"test","test","test/user")
		self.s = self.cdb[uuid.uuid4().hex]
		self.s.create({"type": "boolean"})
	@task
	def addPoint(self):
		self.s.insert(True)
	def __del__(self):
		self.s.delete()



class StreamWriter(BenchLocust):
	task_set = StreamWriteTask
	min_wait = 500
	max_wait = 1500
