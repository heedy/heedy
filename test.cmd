bin\connectordb.exe -l=DEBUG create testdb --sqlbackend=sqlite3
bin\connectordb.exe start testdb
nosetests connectordb-python/connectordb_test.py
nosetests connectordb-python/query_test.py
bin\connectordb.exe stop testdb
