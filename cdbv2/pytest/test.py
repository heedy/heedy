import logging

logging.basicConfig(level=logging.DEBUG)

import grpc

import api_pb2
import api_pb2_grpc

with open("../out/CDBtest.crt", "rb") as f:
    creds = grpc.ssl_channel_credentials(f.read())

with grpc.secure_channel("localhost:3000", creds) as channel:
    stub = api_pb2_grpc.PingStub(channel)
    print("Saying hi")
    msg = api_pb2.PingMessage()
    msg.greeting = "Greetings from Python"
    hi = stub.SayHello(msg)
    print(hi)
