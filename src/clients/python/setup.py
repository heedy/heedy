#!/usr/bin/env python
try:
    from setuptools import setup
except:
    from distutils.core import setup

setup(name='ConnectorDB',
      version='0.2.0',  #The a.b of a.b.c follows connectordb version. c is the version of python. Remember to change __version__ in __init__
      description='ConnectorDB Python Interface',
      author='ConnectorDB team',
      author_email='support@connectordb.com',
      url='http://connectordb.com',
      packages=['connectordb'],
      install_requires=[
          "jsonschema",
          "requests",
          "websocket-client"
          ]
     )