package connectordb

import "connectordb/operator"

// Operator is the base interface used for ConnectorDB
type Operator operator.Operator

// PathOperator is the core interface used for helping query the database in a useful fashion
type PathOperator operator.PathOperator
