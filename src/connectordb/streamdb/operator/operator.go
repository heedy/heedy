package operator

import "connectordb/streamdb/operator/interfaces"

// Operator defines extension functions which work with any BaseOperator, adding
// the ability to query things by path without requiring permission specialization
type Operator interfaces.Operator
