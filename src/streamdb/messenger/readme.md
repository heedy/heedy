Messenger
===========

Allows sending messages containing data and keys. This module is used for immediate actions and low latency downlinks to a device. Ie, when data is inserted into streamdb, it is messaged, so that we can can have low latency pub/sub happen on inserts.

It uses gnatsd internally as of writing.
